package repository

import (
	"database/sql"
	"errors"
	"payment-service/internal/domain"
)

type PostgresPaymentRepository struct {
	DB *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{DB: db}
}

func (r *PostgresPaymentRepository) Create(payment *domain.Payment) error {
	_, err := r.DB.Exec(`
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`, payment.ID, payment.OrderID, payment.TransactionID, payment.Amount, payment.Status)
	return err
}

func (r *PostgresPaymentRepository) GetByOrderID(orderID string) (*domain.Payment, error) {
	row := r.DB.QueryRow(`SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id=$1`, orderID)
	payment := &domain.Payment{}
	err := row.Scan(&payment.ID, &payment.OrderID, &payment.TransactionID, &payment.Amount, &payment.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return payment, nil
}
