package storage

import (
	"fmt"

	"github.com/pluhe7/gophermart/internal/model"
)

type SessionStorage interface {
	Create(session *model.Session) error
	Get(token string) (*model.Session, error)
}

type DatabaseSessionStorage struct {
	Storage
}

func NewSessionStorage(storage Storage) *DatabaseSessionStorage {
	return &DatabaseSessionStorage{
		storage,
	}
}

func (s DatabaseSessionStorage) Create(session *model.Session) error {
	_, err := s.db.Exec("INSERT INTO sessions (token, user_id, created_at, expire_at, last_activity_at) VALUES ($1, $2, $3, $4, $5)",
		session.Token, session.UserID, session.CreatedAt, session.ExpireAt, session.LastActivityAt)
	if err != nil {
		return fmt.Errorf("exec insert: %w", err)
	}

	return nil
}

func (s DatabaseSessionStorage) Get(token string) (*model.Session, error) {
	var session model.Session
	err := s.db.Get(&session, "SELECT * FROM sessions WHERE token=$1", token)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &session, nil
}
