package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pluhe7/gophermart/internal/model"
)

func SessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*Context)

		cookie, err := c.Cookie(model.SessionCookieToken)
		if err == nil && len(cookie.Value) > 0 {
			session, _ := c.App.GetSession(cookie.Value)
			if session != nil {
				c.App.Session = *session
			}

			// update expire if need
		} else {
			return c.String(http.StatusUnauthorized, "wrong cookie")
		}

		if err = next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
