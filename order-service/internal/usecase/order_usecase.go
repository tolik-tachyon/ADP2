package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"order-service/internal/domain"
	"order-service/internal/repository"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	Repo       repository.OrderRepository
	PaymentURL string
	HTTPClient *http.Client
}

func NewOrderUseCase(repo repository.OrderRepository, paymentURL string) *OrderUseCase {
	return &OrderUseCase{
		Repo:       repo,
		PaymentURL: paymentURL,
		HTTPClient: &http.Client{Timeout: 2 * time.Second},
	}
}

func (uc *OrderUseCase) CreateOrder(order *domain.Order, idempotencyKey string) (*domain.Order, error) {
	if idempotencyKey != "" {
		existing, err := uc.Repo.GetByIdempotencyKey(idempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}
	order.ID = uuid.New().String()
	order.Status = "Pending"
	order.CreatedAt = time.Now()
	order.IdempotencyKey = idempotencyKey

	if order.Amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	if err := uc.Repo.Create(order); err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"order_id": order.ID,
		"amount":   order.Amount,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", uc.PaymentURL+"/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := uc.HTTPClient.Do(req)
	if err != nil {
		uc.Repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
		return order, errors.New("payment service unavailable")
	}
	defer resp.Body.Close()

	var res struct {
		Status        string `json:"status"`
		TransactionID string `json:"transaction_id"`
	}
	json.NewDecoder(resp.Body).Decode(&res)

	finalStatus := "Failed"
	if res.Status == "Authorized" {
		finalStatus = "Paid"
	}

	uc.Repo.UpdateStatus(order.ID, finalStatus)
	order.Status = finalStatus
	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.Repo.GetByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) (*domain.Order, error) {
	order, err := uc.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order.Status != "Pending" {
		return nil, errors.New("only pending orders can be cancelled")
	}
	err = uc.Repo.UpdateStatus(id, "Cancelled")
	if err != nil {
		return nil, err
	}
	order.Status = "Cancelled"
	return order, nil
}
