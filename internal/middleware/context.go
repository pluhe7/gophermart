package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/pluhe7/gophermart/internal/app"
)

type Context struct {
	echo.Context

	App *app.App
}

func NewContext(ec echo.Context, a *app.App) *Context {
	cc := &Context{
		Context: ec,
		App:     a,
	}

	return cc
}

func ContextMiddleware(a *app.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ec echo.Context) error {
			c := NewContext(ec, a)
			return next(c)
		}
	}
}
