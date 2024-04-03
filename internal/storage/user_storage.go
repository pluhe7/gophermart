package storage

import (
	"fmt"
	"time"

	"github.com/pluhe7/gophermart/internal/model"
)

type UserStorage interface {
	Create(user *model.User) error
	GetByLogin(login string) (*model.User, error)
	GetBalance(id int) (float64, error)
	UpdateBalance(userID int, sum float64) error
}

type DatabaseUserStorage struct {
	Storage
}

func NewUserStorage(storage Storage) *DatabaseUserStorage {
	return &DatabaseUserStorage{
		storage,
	}
}

func (s DatabaseUserStorage) Create(user *model.User) error {
	user.CreatedAt = time.Now()

	_, err := s.db.Exec("INSERT INTO users (login, password_hash, created_at) VALUES ($1, $2, $3)",
		user.Login, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}

	return nil
}

func (s DatabaseUserStorage) GetByLogin(login string) (*model.User, error) {
	var user model.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE login=$1", login)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &user, nil
}

func (s DatabaseUserStorage) GetBalance(id int) (float64, error) {
	var balance float64
	err := s.db.Get(&balance, "SELECT balance FROM users WHERE id=$1", id)
	if err != nil {
		return 0, fmt.Errorf("get: %w", err)
	}

	return balance, nil
}

func (s DatabaseUserStorage) UpdateBalance(userID int, sum float64) error {
	_, err := s.db.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2",
		sum, userID)
	if err != nil {
		return fmt.Errorf("exec update: %w", err)
	}

	return nil
}
