package storage

import (
	"fmt"
	"time"

	"github.com/pluhe7/gophermart/internal/model"
)

type OrderStorage interface {
	Create(order *model.Order) error
	Get(number string) (*model.Order, error)
	UpdateStatus(number string, status model.OrderStatus) error
	FindUnprocessed() ([]*model.Order, error)
	FindForUser(userID int) ([]model.Order, error)
}

type DatabaseOrderStorage struct {
	Storage
}

func NewOrderStorage(storage Storage) *DatabaseOrderStorage {
	return &DatabaseOrderStorage{
		storage,
	}
}

func (s DatabaseOrderStorage) Create(order *model.Order) error {
	order.Status = model.OrderStatusNew
	order.UploadedAt = time.Now()

	_, err := s.db.Exec("INSERT INTO orders (number, user_id, status, uploaded_at) VALUES ($1, $2, $3, $4)",
		order.Number, order.UserID, order.Status, order.UploadedAt)
	if err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}

	return nil
}

func (s DatabaseOrderStorage) Get(number string) (*model.Order, error) {
	var order model.Order
	err := s.db.Get(&order, "SELECT * FROM orders WHERE number=$1", number)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &order, nil
}

func (s DatabaseOrderStorage) UpdateStatus(number string, status model.OrderStatus) error {
	_, err := s.db.Exec("UPDATE orders SET status=$1 WHERE number=$2", status, number)
	if err != nil {
		return fmt.Errorf("exec update: %w", err)
	}

	return nil
}

func (s DatabaseOrderStorage) FindUnprocessed() ([]*model.Order, error) {
	var orders []*model.Order
	err := s.db.Select(&orders, "SELECT * FROM orders WHERE status IN($1, $2)",
		model.OrderStatusNew, model.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return orders, nil
}

func (s DatabaseOrderStorage) FindForUser(userID int) ([]model.Order, error) {
	var orders []model.Order
	err := s.db.Select(&orders, `SELECT o.*, t.sum AS accrual FROM orders o 
		LEFT JOIN transactions t ON t.order_number = o.number AND t.type = $1 
		WHERE o.user_id = $2`,
		model.TransactionTypeAccrual, userID)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return orders, nil
}
