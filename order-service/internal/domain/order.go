package domain

import "time"

const (
	OrderStatusPending   = "Pending"
	OrderStatusPaid      = "Paid"
	OrderStatusFailed    = "Failed"
	OrderStatusCancelled = "Cancelled"
)

type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
