package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"order-service/internal/domain"
	"order-service/internal/repository"

	pb "github.com/tolik-tachyon/proto-generated/paymentpb"

	"order-service/internal/cache"
)

type OrderUseCase struct {
	Repo          repository.OrderRepository
	PaymentClient pb.PaymentServiceClient
	Cache         *cache.RedisCache
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

	// cache newly created order
	go func() {
		if uc.Cache != nil {
			b, _ := json.Marshal(order)
			_ = uc.Cache.Set(context.Background(), "order:"+order.ID, string(b), 5*time.Minute)
		}
	}()

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

	// invalidate cache after status change
	if uc.Cache != nil {
		_ = uc.Cache.Delete(context.Background(), "order:"+order.ID)
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	// try cache first
	if uc.Cache != nil {
		if val, err := uc.Cache.Get(context.Background(), "order:"+id); err == nil {
			var o domain.Order
			if json.Unmarshal([]byte(val), &o) == nil {
				return &o, nil
			}
		}
	}

	order, err := uc.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// store in cache
	if uc.Cache != nil {
		b, _ := json.Marshal(order)
		_ = uc.Cache.Set(context.Background(), "order:"+id, string(b), 5*time.Minute)
	}

	return order, nil
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

	// invalidate cache
	if uc.Cache != nil {
		_ = uc.Cache.Delete(context.Background(), "order:"+id)
	}

	return order, nil
}
