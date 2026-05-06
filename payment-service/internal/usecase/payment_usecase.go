package usecase

import (
	"payment-service/internal/domain"
	"payment-service/internal/messaging"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	Repo      repository.PaymentRepository
	Publisher *messaging.Publisher
}

func NewPaymentUseCase(repo repository.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{Repo: repo}
}

func (uc *PaymentUseCase) AuthorizePayment(payment *domain.Payment) error {
	payment.ID = uuid.New().String()

	if payment.Amount > 100000 {
		payment.Status = "Declined"
		payment.TransactionID = ""
	} else {
		payment.Status = "Authorized"
		payment.TransactionID = uuid.New().String()
	}

	if err := uc.Repo.Create(payment); err != nil {
		return err
	}

	if payment.Status == "Authorized" && uc.Publisher != nil {
		err := uc.Publisher.PublishPaymentCompleted(messaging.PaymentEvent{
			EventID:       uuid.New().String(),
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: "user@example.com",
			Status:        payment.Status,
		})

		if err != nil {
			return nil
		}
	}

	return nil
}

func (uc *PaymentUseCase) GetPayment(orderID string) (*domain.Payment, error) {
	return uc.Repo.GetByOrderID(orderID)
}

func (uc *PaymentUseCase) ListPayments(status string) ([]*domain.Payment, error) {
	return uc.Repo.ListByStatus(status)
}
