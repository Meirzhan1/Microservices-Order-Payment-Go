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

func (r *PostgresPaymentRepository) FindByAmountRange(ctx context.Context, min, max int64) ([]*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status, decline_reason, created_at
		FROM payments
		WHERE 1=1
	`
	var args []interface{}
	argId := 1

	if min > 0 {
		query += ` AND amount >= $` + itoa(argId)
		args = append(args, min)
		argId++
	}
	if max > 0 {
		query += ` AND amount <= $` + itoa(argId)
		args = append(args, max)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &p.DeclineReason, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func itoa(i int) string {
	var b [32]byte
	bp := len(b)
	for i > 0 || bp == len(b) {
		bp--
		b[bp] = byte(i%10) + '0'
		i /= 10
	}
	return string(b[bp:])
}
