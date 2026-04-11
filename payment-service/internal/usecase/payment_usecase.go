package usecase

import (
	"context"
	"errors"
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
}

type CreatePaymentInput struct {
	OrderID string
	Amount  int64
}

type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(repo PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
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

func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, input CreatePaymentInput) (*ProcessPaymentResult, error) {
	payment, err := uc.CreatePayment(ctx, input)
	if err != nil {
		return nil, err
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
