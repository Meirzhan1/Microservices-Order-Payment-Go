package usecase

import (
	"context"
	"errors"
	"fmt"
	"order-service/internal/domain"
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

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type PaymentClient interface {
	CreatePayment(ctx context.Context, orderID string, amount int64, customerEmail string) (string, error)
}

type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
	cache         Cache
	cacheTTL      time.Duration
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient, cache Cache, ttl time.Duration) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentClient: paymentClient, cache: cache, cacheTTL: ttl}
}

type CreateOrderInput struct {
	CustomerID     string
	CustomerEmail  string
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
		ID:            uuid.NewString(),
		CustomerID:    input.CustomerID,
		CustomerEmail: input.CustomerEmail,
		ItemName:      input.ItemName,
		Amount:        input.Amount,
		Status:        domain.OrderStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := uc.repo.Create(ctx, order, input.IdempotencyKey); err != nil {
		return nil, err
	}

	paymentStatus, err := uc.paymentClient.CreatePayment(ctx, order.ID, order.Amount, order.CustomerEmail)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed, time.Now().UTC())
		order.Status = domain.OrderStatusFailed
		order.UpdatedAt = time.Now().UTC()
		return order, ErrPaymentUnavailable
	}

	if paymentStatus == "Authorized" {
		order.Status = domain.OrderStatusPaid
	} else {
		order.Status = domain.OrderStatusFailed
	}
	order.UpdatedAt = time.Now().UTC()

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status, order.UpdatedAt); err != nil {
		return nil, err
	}

	uc.invalidateOrderCache(ctx, order.ID)

	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	var order domain.Order
	cacheKey := fmt.Sprintf("order:%s", id)

	if err := uc.cache.Get(ctx, cacheKey, &order); err == nil && order.ID != "" {
		return &order, nil
	}

	dbOrder, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dbOrder == nil {
		return nil, ErrOrderNotFound
	}

	_ = uc.cache.Set(ctx, cacheKey, dbOrder, uc.cacheTTL)

	return dbOrder, nil
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

	uc.invalidateOrderCache(ctx, id)

	return order, nil
}

func (uc *OrderUseCase) invalidateOrderCache(ctx context.Context, id string) {
	cacheKey := fmt.Sprintf("order:%s", id)
	_ = uc.cache.Delete(ctx, cacheKey)
}
