package repository

import (
	"database/sql"
	"errors"
	"order-service/internal/domain"
)

type PostgresOrderRepository struct {
	DB *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{DB: db}
}

func (r *PostgresOrderRepository) Create(order *domain.Order) error {
	_, err := r.DB.Exec(`
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, order.CreatedAt, order.IdempotencyKey)
	return err
}

func (r *PostgresOrderRepository) UpdateStatus(id string, status string) error {
	res, err := r.DB.Exec(`UPDATE orders SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *PostgresOrderRepository) GetByID(id string) (*domain.Order, error) {
	row := r.DB.QueryRow(`SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE id=$1`, id)
	order := &domain.Order{}
	err := row.Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt, &order.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *PostgresOrderRepository) GetByIdempotencyKey(key string) (*domain.Order, error) {
	if key == "" {
		return nil, nil
	}
	row := r.DB.QueryRow(`SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE idempotency_key=$1`, key)
	order := &domain.Order{}
	err := row.Scan(&order.ID, &order.CustomerID, &order.ItemName, &order.Amount, &order.Status, &order.CreatedAt, &order.IdempotencyKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}
