package middleware

import (
	"io"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/logger"
)

func RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*Context)

		start := time.Now()

		if err := next(c); err != nil {
			c.Error(err)
		}

		duration := time.Since(start)

		reqBody, _ := io.ReadAll(c.Request().Body)

		logger.Log.Info("got incoming HTTP request",
			zap.Duration("duration", duration),
			zap.String("url", c.Request().URL.Path),
			zap.String("req", string(reqBody)),
			zap.Int("status", c.Response().Status),
			zap.Int64("resp size", c.Response().Size),
		)

		return nil
	}
}
