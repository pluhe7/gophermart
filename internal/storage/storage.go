package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/logger"
)

const (
	DBPingAttempts    = 10
	DBPingTimeoutSecs = 10
)

type Storage struct {
	db  *sqlx.DB
	dsn string

	UserStorage        UserStorage
	SessionStorage     SessionStorage
	OrderStorage       OrderStorage
	TransactionStorage TransactionStorage
}

func NewStorage(dsn string) (*Storage, error) {
	storage := &Storage{
		dsn: dsn,
	}

	err := storage.initConnection()
	if err != nil {
		return nil, fmt.Errorf("init connection: %w", err)
	}

	storage.UserStorage = NewUserStorage(*storage)
	storage.SessionStorage = NewSessionStorage(*storage)
	storage.OrderStorage = NewOrderStorage(*storage)
	storage.TransactionStorage = NewTransactionStorage(*storage)

	err = storage.initMigrations()
	if err != nil {
		return nil, fmt.Errorf("init migrations: %w", err)
	}

	return storage, err
}

func (s *Storage) initConnection() error {
	db, err := sqlx.Open("postgres", s.dsn)
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	for i := 0; i < DBPingAttempts; i++ {
		logger.Log.Info("pinging database")

		ctx, cancel := context.WithTimeout(context.Background(), DBPingTimeoutSecs*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			if i == DBPingAttempts-1 {
				return fmt.Errorf("ping database: %w", err)
			} else {
				logger.Log.Error(fmt.Sprintf("ping database error, retrying in %d seconds", DBPingTimeoutSecs), zap.Error(err))
				time.Sleep(DBPingTimeoutSecs * time.Second)
			}
		} else {
			break
		}
	}

	s.db = db

	return err
}

func (s *Storage) initMigrations() error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	err = goose.Up(s.db.DB, "internal/storage/migrations", goose.WithAllowMissing())
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return err
}

func (s *Storage) DB() *sqlx.DB {
	return s.db
}

func (s *Storage) Close() error {
	logger.Log.Info("closing storage database")
	err := s.DB().Close()
	if err != nil {
		return fmt.Errorf("closing storage database: %w", err)
	}

	return nil
}

func (s *Storage) User() UserStorage {
	return s.UserStorage
}

func (s *Storage) Session() SessionStorage {
	return s.SessionStorage
}

func (s *Storage) Order() OrderStorage {
	return s.OrderStorage
}

func (s *Storage) Transaction() TransactionStorage {
	return s.TransactionStorage
}
