package model

import "time"

type OrderStatus string

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string      `json:"number" db:"number"`
	UserID     int         `json:"user_id" db:"user_id"`
	Status     OrderStatus `json:"status" db:"status"`
	UploadedAt time.Time   `json:"uploaded_at" db:"uploaded_at"`

	Accrual *float64 `json:"accrual,omitempty" db:"accrual"`
}
