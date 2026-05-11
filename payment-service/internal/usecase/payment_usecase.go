package usecase

import (
	"context"
	"errors"
	"log"
	"payment-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidAmount = errors.New("amount must be > 0")

type ProcessPaymentResult struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	DeclineReason string
	CreatedAt     time.Time
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
	FindByAmountRange(ctx context.Context, min, max int64) ([]*domain.Payment, error)
}

type PaymentEvent struct {
	PaymentID     string `json:"payment_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type PaymentEventProducer interface {
	PublishPaymentCompleted(ctx context.Context, event PaymentEvent) error
}

type CreatePaymentInput struct {
	OrderID string
	Amount  int64
}

type PaymentUseCase struct {
	repo     PaymentRepository
	producer PaymentEventProducer
}

func NewPaymentUseCase(repo PaymentRepository, producer PaymentEventProducer) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, producer: producer}
}

func (uc *PaymentUseCase) CreatePayment(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	existing, err := uc.repo.GetByOrderID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	payment := &domain.Payment{
		ID:            uuid.NewString(),
		OrderID:       input.OrderID,
		Amount:        input.Amount,
		CreatedAt:     time.Now().UTC(),
		TransactionID: uuid.NewString(),
	}

	if input.Amount > domain.PaymentLimitCents {
		payment.Status = domain.PaymentStatusDeclined
		payment.DeclineReason = "amount exceeds limit"
	} else {
		payment.Status = domain.PaymentStatusAuthorized
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}

func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, input CreatePaymentInput, customerEmail string) (*ProcessPaymentResult, error) {
	payment, err := uc.CreatePayment(ctx, input)
	if err != nil {
		return nil, err
	}

	if payment.Status == domain.PaymentStatusAuthorized {
		event := PaymentEvent{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: customerEmail,
			Status:        payment.Status,
		}
		if uc.producer != nil {
			if err := uc.producer.PublishPaymentCompleted(ctx, event); err != nil {
				log.Printf("[CRITICAL] Failed to publish payment completed event for Order #%s: %v", payment.OrderID, err)
				return nil, errors.New("payment processed but failed to trigger notification (broker error)")
			}
		}
	}

	return &ProcessPaymentResult{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		DeclineReason: payment.DeclineReason,
		CreatedAt:     payment.CreatedAt,
	}, nil
}

func (uc *PaymentUseCase) ListPayments(ctx context.Context, min, max int64) ([]*ProcessPaymentResult, error) {
	if min > 0 && max > 0 && min > max {
		return nil, errors.New("min_amount cannot be greater than max_amount")
	}

	payments, err := uc.repo.FindByAmountRange(ctx, min, max)
	if err != nil {
		return nil, err
	}

	var results []*ProcessPaymentResult
	for _, p := range payments {
		results = append(results, &ProcessPaymentResult{
			ID:            p.ID,
			OrderID:       p.OrderID,
			TransactionID: p.TransactionID,
			Amount:        p.Amount,
			Status:        p.Status,
			DeclineReason: p.DeclineReason,
			CreatedAt:     p.CreatedAt,
		})
	}
	return results, nil
}
