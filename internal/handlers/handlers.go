package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pluhe7/gophermart/internal/app"
	"github.com/pluhe7/gophermart/internal/middleware"
	"github.com/pluhe7/gophermart/internal/model"
)

func InitHandlers(a *app.App) {
	a.Server.Echo.Use(middleware.ContextMiddleware(a),
		middleware.RequestLoggerMiddleware)

	userGroup := a.Server.Echo.Group("/api/user")

	userGroup.POST("/register", echoHandler(registerHandler))
	userGroup.POST("/login", echoHandler(loginHandler))

	userGroupCopy := *userGroup
	sessionUserGroup := &userGroupCopy

	sessionUserGroup.Use(middleware.SessionMiddleware)

	sessionUserGroup.GET("/orders", echoHandler(getOrdersHandler))
	sessionUserGroup.POST("/orders", echoHandler(sendOrderHandler))

	sessionUserGroup.GET("/balance", echoHandler(getBalanceHandler))
	sessionUserGroup.POST("/balance/withdraw", echoHandler(withdrawHandler))

	sessionUserGroup.GET("/withdrawals", echoHandler(getWithdrawalsHandler))
}

func echoHandler(h func(c *middleware.Context) error) echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*middleware.Context)
		return h(c)
	}
}

func registerHandler(c *middleware.Context) error {
	var userData model.UserLoginData
	err := c.Bind(&userData)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = c.App.CreateUser(userData)
	if err != nil {
		if errors.Is(err, model.ErrUserExist) {
			return c.String(http.StatusConflict, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	session, err := c.App.Login(userData)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	sessionCookie := &http.Cookie{
		Name:    model.SessionCookieToken,
		Value:   session.Token,
		Expires: session.ExpireAt,
	}

	c.SetCookie(sessionCookie)

	return c.NoContent(http.StatusOK)
}

func loginHandler(c *middleware.Context) error {
	var userData model.UserLoginData
	err := c.Bind(&userData)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	session, err := c.App.Login(userData)
	if err != nil {
		if errors.Is(err, model.ErrUserNotExist) {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	sessionCookie := &http.Cookie{
		Name:    model.SessionCookieToken,
		Value:   session.Token,
		Expires: session.ExpireAt,
	}

	c.SetCookie(sessionCookie)

	return c.NoContent(http.StatusOK)
}

func getOrdersHandler(c *middleware.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	orders, err := c.App.GetOrders()
	if err != nil {
		if errors.Is(err, model.ErrNoOrders) {
			return c.NoContent(http.StatusNoContent)
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, orders)
}

func sendOrderHandler(c *middleware.Context) error {
	orderNumberBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = c.App.SendOrder(string(orderNumberBytes))
	if err != nil {
		if errors.Is(err, model.ErrCurrentUserOrderExist) {
			return c.String(http.StatusOK, err.Error())
		}

		if errors.Is(err, model.ErrOtherUserOrderExist) {
			return c.String(http.StatusConflict, err.Error())
		}

		if errors.Is(err, model.ErrOrderNumberFormat) {
			return c.String(http.StatusUnprocessableEntity, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusAccepted)
}

func getBalanceHandler(c *middleware.Context) error {
	balance, err := c.App.GetBalance()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, balance)
}

func withdrawHandler(c *middleware.Context) error {
	var withdrawalData model.WithdrawalData
	err := c.Bind(&withdrawalData)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = c.App.Withdraw(withdrawalData)
	if err != nil {
		if errors.Is(err, model.ErrLowBalance) {
			return c.String(http.StatusPaymentRequired, err.Error())
		}

		if errors.Is(err, model.ErrWrongOrderNumber) {
			return c.String(http.StatusUnprocessableEntity, err.Error())
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func getWithdrawalsHandler(c *middleware.Context) error {
	withdrawals, err := c.App.GetWithdrawals()
	if err != nil {
		if errors.Is(err, model.ErrNoWithdrawals) {
			return c.NoContent(http.StatusNoContent)
		}

		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, withdrawals)
}
