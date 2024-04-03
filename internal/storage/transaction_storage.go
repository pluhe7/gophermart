package storage

import (
	"fmt"
	"time"

	"github.com/pluhe7/gophermart/internal/model"
)

type TransactionStorage interface {
	Create(transaction *model.Transaction) error
	FindWithdrawsForUser(userID int) ([]*model.Transaction, error)
}

type DatabaseTransactionStorage struct {
	Storage
}

func NewTransactionStorage(storage Storage) *DatabaseTransactionStorage {
	return &DatabaseTransactionStorage{
		storage,
	}
}

func (s DatabaseTransactionStorage) Create(transaction *model.Transaction) error {
	transaction.ProcessedAt = time.Now()

	_, err := s.db.Exec("INSERT INTO transactions (user_id, order_number, sum, type, processed_at) VALUES ($1, $2, $3, $4, $5)",
		transaction.UserID, transaction.OrderNumber, transaction.Sum, transaction.Type, transaction.ProcessedAt)
	if err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}

	return nil
}

func (s DatabaseTransactionStorage) FindWithdrawsForUser(userID int) ([]*model.Transaction, error) {
	var withdraws []*model.Transaction
	err := s.db.Select(&withdraws, "SELECT * FROM transactions WHERE user_id=$1 AND type=$2",
		userID, model.TransactionTypeWithdrawal)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return withdraws, nil
}
