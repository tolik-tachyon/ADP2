package usecase

import (
	"payment-service/internal/domain"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	Repo repository.PaymentRepository
}

func NewPaymentUseCase(repo repository.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{Repo: repo}
}

func (uc *PaymentUseCase) AuthorizePayment(payment *domain.Payment) error {
	payment.ID = uuid.New().String()

	// Проверка лимита
	if payment.Amount > 100000 {
		payment.Status = "Declined"
		payment.TransactionID = ""
	} else {
		payment.Status = "Authorized"
		payment.TransactionID = uuid.New().String()
	}

	return uc.Repo.Create(payment)
}

func (uc *PaymentUseCase) GetPayment(orderID string) (*domain.Payment, error) {
	return uc.Repo.GetByOrderID(orderID)
}
