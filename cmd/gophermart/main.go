package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/app"
	"github.com/pluhe7/gophermart/internal/config"
	"github.com/pluhe7/gophermart/internal/handlers"
	"github.com/pluhe7/gophermart/internal/logger"
)

func main() {
	cfg := config.InitConfig()

	logger.InitLogger(cfg.LogLevel)

	server, err := app.NewServer(cfg)
	if err != nil {
		logger.Log.Fatal("creating server", zap.Error(err))
	}

	a := app.NewApp(server)

	ctxSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	a.ProcessOrders(ctxSignal)
	a.UpdateOrders(ctxSignal)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.StartOrderProcessor(ctxSignal)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		handlers.InitHandlers(a)

		err = server.Start()
		if err != nil {
			logger.Log.Fatal("starting server", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		<-ctxSignal.Done()

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer func() {
			wg.Done()
			stop()
			cancel()
		}()

		if err = server.Stop(ctxTimeout); err != nil {
			logger.Log.Error("error stopping server", zap.Error(err))
			return
		}
	}()

	wg.Wait()
}
