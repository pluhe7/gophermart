package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/config"
	"github.com/pluhe7/gophermart/internal/logger"
	"github.com/pluhe7/gophermart/internal/storage"
)

type Server struct {
	Config         *config.Config
	Storage        *storage.Storage
	Echo           *echo.Echo
	OrderProcessor *OrderProcessor
}

func NewServer(cfg *config.Config) (*Server, error) {
	s, err := storage.NewStorage(cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("create new storage: %w", err)
	}

	server := &Server{
		Storage:        s,
		Config:         cfg,
		Echo:           echo.New(),
		OrderProcessor: NewOrderProcessor(cfg.AccrualSystemAddress),
	}

	return server, nil
}

func (s *Server) Start() error {
	logger.Log.Info("starting server...", zap.Object("config", s.Config))

	err := s.Echo.Start(s.Config.Address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Fatal("starting server error", zap.Error(err))
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	logger.Log.Info("stopping server...")

	if s.Echo != nil {
		if err := s.Echo.Shutdown(ctx); err != nil {
			return fmt.Errorf("echo shutdown: %w", err)
		}

		s.Echo = nil
	}

	if s.Storage != nil {
		if err := s.Storage.Close(); err != nil {
			return fmt.Errorf("close storage: %w", err)
		}
	}

	logger.Log.Info("server stopped")
	return nil
}
