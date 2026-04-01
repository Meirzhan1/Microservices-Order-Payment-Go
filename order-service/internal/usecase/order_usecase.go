package usecase

import (
	"context"
	"errors"
	"order-service/internal/domain"
	"order-service/internal/transport/httpclient"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount          = errors.New("amount must be > 0")
	ErrOrderNotFound          = errors.New("order not found")
	ErrOrderCannotBeCancelled = errors.New("only pending orders can be cancelled")
	ErrPaymentUnavailable     = errors.New("payment service unavailable")
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order, idempotencyKey string) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string, updatedAt time.Time) error
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.Order, error)
}

type PaymentClient interface {
	CreatePayment(ctx context.Context, orderID string, amount int64) (*httpclient.CreatePaymentResponse, error)
}

type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentClient: paymentClient}
}

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if input.IdempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	now := time.Now().UTC()
	order := &domain.Order{
		ID:         uuid.NewString(),
		CustomerID: input.CustomerID,
		ItemName:   input.ItemName,
		Amount:     input.Amount,
		Status:     domain.OrderStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.repo.Create(ctx, order, input.IdempotencyKey); err != nil {
		return nil, err
	}

	paymentResp, err := uc.paymentClient.CreatePayment(ctx, order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed, time.Now().UTC())
		order.Status = domain.OrderStatusFailed
		order.UpdatedAt = time.Now().UTC()
		return order, ErrPaymentUnavailable
	}

	if paymentResp.Status == "Authorized" {
		order.Status = domain.OrderStatusPaid
	} else {
		order.Status = domain.OrderStatusFailed
	}
	order.UpdatedAt = time.Now().UTC()

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status, order.UpdatedAt); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	if order.Status != domain.OrderStatusPending {
		return nil, ErrOrderCannotBeCancelled
	}

	now := time.Now().UTC()
	if err := uc.repo.UpdateStatus(ctx, id, domain.OrderStatusCancelled, now); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = now
	return order, nil
}
