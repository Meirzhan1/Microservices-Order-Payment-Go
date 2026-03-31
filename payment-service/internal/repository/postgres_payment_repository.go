package repository

import (
	"context"
	"database/sql"
	"errors"
	"payment-service/internal/domain"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	const query = `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, decline_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
		payment.DeclineReason,
		payment.CreatedAt,
	)

	return err
}

func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	const query = `
		SELECT id, order_id, transaction_id, amount, status, decline_reason, created_at
		FROM payments
		WHERE order_id = $1
	`

	payment := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
		&payment.DeclineReason,
		&payment.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return payment, nil
}
