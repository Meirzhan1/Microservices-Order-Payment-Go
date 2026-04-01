package repository

import (
	"context"
	"database/sql"
	"errors"
	"order-service/internal/domain"
	"time"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order, idempotencyKey string) error {
	const query = `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, updated_at, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
		idempotencyKey,
	)

	return err
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	const query = `
		SELECT id, customer_id, item_name, amount, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status string, updatedAt time.Time) error {
	const query = `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, updatedAt, id)
	return err
}

func (r *PostgresOrderRepository) GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.Order, error) {
	const query = `
		SELECT id, customer_id, item_name, amount, status, created_at, updated_at
		FROM orders
		WHERE idempotency_key = $1
	`

	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, query, idempotencyKey).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}
