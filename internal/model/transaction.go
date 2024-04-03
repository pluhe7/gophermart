package model

import "time"

type TransactionType string

const (
	TransactionTypeWithdrawal = "withdrawal"
	TransactionTypeAccrual    = "accrual"
)

type Transaction struct {
	ID          int             `json:"id" db:"id"`
	UserID      int             `json:"user_id" db:"user_id"`
	OrderNumber string          `json:"order_number" db:"order_number"`
	Sum         float64         `json:"sum" db:"sum"`
	Type        TransactionType `json:"type" db:"type"`
	ProcessedAt time.Time       `json:"processed_at" db:"processed_at"`
}
