package domain

import "time"

const (
	PaymentStatusAuthorized       = "Authorized"
	PaymentStatusDeclined         = "Declined"
	PaymentLimitCents       int64 = 100000
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
	DeclineReason string
	CreatedAt     time.Time
}
