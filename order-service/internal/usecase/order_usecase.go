package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"order-service/internal/domain"
	"order-service/internal/repository"

	pb "github.com/tolik-tachyon/proto-generated/paymentpb"
)

type OrderUseCase struct {
	Repo          repository.OrderRepository
	PaymentClient pb.PaymentServiceClient
}

func NewOrderUseCase(repo repository.OrderRepository, client pb.PaymentServiceClient) *OrderUseCase {
	return &OrderUseCase{
		Repo:          repo,
		PaymentClient: client,
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := uc.PaymentClient.ProcessPayment(
		ctx,
		&pb.PaymentRequest{
			OrderId: order.ID,
			Amount:  order.Amount,
		},
	)

	if err != nil {
		uc.Repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
		return order, errors.New("payment service unavailable")
	}

	finalStatus := "Failed"
	if resp.Status == "Authorized" {
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
