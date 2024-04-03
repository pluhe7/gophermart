package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/pluhe7/gophermart/internal/app"
	"github.com/pluhe7/gophermart/internal/config"
	"github.com/pluhe7/gophermart/internal/middleware"
	"github.com/pluhe7/gophermart/internal/mocks"
	"github.com/pluhe7/gophermart/internal/model"
	"github.com/pluhe7/gophermart/internal/storage"
)

func testServer(cfg *config.Config, mockController *gomock.Controller) *app.Server {
	s := &storage.Storage{}

	s.UserStorage = mocks.NewMockUserStorage(mockController)
	s.SessionStorage = mocks.NewMockSessionStorage(mockController)
	s.OrderStorage = mocks.NewMockOrderStorage(mockController)
	s.TransactionStorage = mocks.NewMockTransactionStorage(mockController)

	return &app.Server{
		Storage:        s,
		Config:         cfg,
		Echo:           echo.New(),
		OrderProcessor: app.NewOrderProcessor(cfg.AccrualSystemAddress),
	}
}

func TestRegisterHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)

	cc := &middleware.Context{
		App: app.NewApp(srv),
	}

	type want struct {
		statusCode int
		resp       string
	}

	tests := []struct {
		name string
		req  string
		want want
	}{
		{
			name: "success",
			req:  `{"login":"some_login", "password":"SoMepASSword1"}`,
			want: want{
				statusCode: http.StatusOK,
				resp:       "",
			},
		},
		{
			name: "duplicate user",
			req:  `{"login":"some_login", "password":"SoMepASSword1"}`,
			want: want{
				statusCode: http.StatusConflict,
				resp:       "user already exist",
			},
		},
		{
			name: "internal error",
			req:  `{"login":"some_login", "password":"SoMepASSword1"}`,
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "create user: some error",
			},
		},
		{
			name: "wrong request",
			req:  `not expected body`,
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "code=400, message=Syntax error: offset=2, error=invalid character 'o' in literal null (expecting 'u'), internal=invalid character 'o' in literal null (expecting 'u')",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusOK:
				firstGetByLoginMock := srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(nil, sql.ErrNoRows)
				secondGetByLoginMock := srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(&model.User{
					ID:           1,
					Login:        "some_login",
					PasswordHash: "$2a$10$GTNDyDXvaok4Qee7IhoSXO2dW6sxTy3GQwmXQqqt4CsPY9cf9T1Dq",
				}, nil)

				gomock.InOrder(
					firstGetByLoginMock,
					secondGetByLoginMock,
				)

				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().Create(gomock.Any()).Return(nil)
				srv.Storage.Session().(*mocks.MockSessionStorage).EXPECT().Create(gomock.Any()).Return(nil)

			case http.StatusConflict:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(&model.User{
					ID:           1,
					Login:        "some_login",
					PasswordHash: "$2a$10$GTNDyDXvaok4Qee7IhoSXO2dW6sxTy3GQwmXQqqt4CsPY9cf9T1Dq",
				}, nil)

			case http.StatusInternalServerError:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(nil, sql.ErrNoRows)
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().Create(gomock.Any()).Return(errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(test.req))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			registerHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			require.Equal(t, test.want.resp, string(resultBody))
		})
	}
}

func TestLoginHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)

	cc := &middleware.Context{
		App: app.NewApp(srv),
	}

	type want struct {
		statusCode int
		resp       string
	}

	tests := []struct {
		name string
		req  string
		want want
	}{
		{
			name: "success",
			req:  `{"login":"some_login", "password":"SoMepASSword1"}`,
			want: want{
				statusCode: http.StatusOK,
				resp:       "",
			},
		},
		{
			name: "not registered user",
			req:  `{"login":"some_login", "password":"SoMepASSword1"}`,
			want: want{
				statusCode: http.StatusUnauthorized,
				resp:       "user not exist",
			},
		},
		{
			name: "internal error",
			req:  `{"login":"some_login", "password":"SoMepASSword1"}`,
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "get user: some error",
			},
		},
		{
			name: "wrong request",
			req:  `not expected body`,
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "code=400, message=Syntax error: offset=2, error=invalid character 'o' in literal null (expecting 'u'), internal=invalid character 'o' in literal null (expecting 'u')",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusOK:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(&model.User{
					ID:           1,
					Login:        "some_login",
					PasswordHash: "$2a$10$GTNDyDXvaok4Qee7IhoSXO2dW6sxTy3GQwmXQqqt4CsPY9cf9T1Dq",
				}, nil)
				srv.Storage.Session().(*mocks.MockSessionStorage).EXPECT().Create(gomock.Any()).Return(nil)

			case http.StatusUnauthorized:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(nil, sql.ErrNoRows)

			case http.StatusInternalServerError:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetByLogin(gomock.Any()).Return(nil, errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(test.req))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			loginHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			require.Equal(t, test.want.resp, string(resultBody))
		})
	}
}

func TestGetOrdersHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)
	a := app.NewApp(srv)
	a.Session = model.Session{
		Token:  "token",
		UserID: 1,
	}

	cc := &middleware.Context{
		App: a,
	}

	orders := []model.Order{
		{
			Number:  "9278923470",
			UserID:  1,
			Status:  model.OrderStatusProcessed,
			Accrual: nil,
		},
		{
			Number:  "12345678903",
			UserID:  1,
			Status:  model.OrderStatusProcessing,
			Accrual: nil,
		},
	}

	successOrdersResp, err := json.Marshal(orders)
	require.NoError(t, err)

	type want struct {
		statusCode int
		resp       string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "success",
			want: want{
				statusCode: http.StatusOK,
				resp:       string(successOrdersResp),
			},
		},
		{
			name: "no orders",
			want: want{
				statusCode: http.StatusNoContent,
				resp:       "",
			},
		},
		{
			name: "internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "find orders for user: some error",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusOK:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().FindForUser(1).Return(orders, nil)

			case http.StatusNoContent:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().FindForUser(1).Return(nil, sql.ErrNoRows)

			case http.StatusInternalServerError:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().FindForUser(1).Return(nil, errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodGet, "/user/orders", nil)
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			getOrdersHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			if test.want.statusCode == http.StatusOK {
				require.JSONEq(t, test.want.resp, string(resultBody))
			} else {
				require.Equal(t, test.want.resp, string(resultBody))
			}
		})
	}
}

func TestSendOrderHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)
	a := app.NewApp(srv)
	a.Session = model.Session{
		Token:  "token",
		UserID: 1,
	}

	cc := &middleware.Context{
		App: a,
	}

	type want struct {
		statusCode int
		resp       string
	}

	tests := []struct {
		name string
		req  string
		want want
	}{
		{
			name: "success",
			req:  "12345678903",
			want: want{
				statusCode: http.StatusAccepted,
				resp:       "",
			},
		},
		{
			name: "order exist",
			req:  "12345678903",
			want: want{
				statusCode: http.StatusOK,
				resp:       "order already was sent",
			},
		},
		{
			name: "other user order",
			req:  "12345678903",
			want: want{
				statusCode: http.StatusConflict,
				resp:       "other user already sent this order",
			},
		},
		{
			name: "wrong number",
			req:  "222222",
			want: want{
				statusCode: http.StatusUnprocessableEntity,
				resp:       "wrong order number",
			},
		},
		{
			name: "internal error",
			req:  "12345678903",
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "get existing order: some error",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusAccepted:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().Get(test.req).Return(nil, sql.ErrNoRows)
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().Create(gomock.Any()).Return(nil)

			case http.StatusOK:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().Get(test.req).Return(&model.Order{
					Number: test.req,
					UserID: 1,
					Status: model.OrderStatusProcessing,
				}, nil)

			case http.StatusConflict:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().Get(test.req).Return(&model.Order{
					Number: test.req,
					UserID: 22,
					Status: model.OrderStatusProcessing,
				}, nil)

			case http.StatusInternalServerError:
				srv.Storage.Order().(*mocks.MockOrderStorage).EXPECT().Get(test.req).Return(nil, errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodPost, "/user/orders", strings.NewReader(test.req))
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			sendOrderHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			require.Equal(t, test.want.resp, string(resultBody))
		})
	}
}

func TestGetBalanceHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)
	a := app.NewApp(srv)
	a.Session = model.Session{
		Token:  "token",
		UserID: 1,
	}

	cc := &middleware.Context{
		App: a,
	}

	type want struct {
		statusCode int
		resp       string
	}

	balance := model.Balance{
		Current:   1234.56,
		Withdrawn: 133.33,
	}

	successResp, err := json.Marshal(balance)
	require.NoError(t, err)

	tests := []struct {
		name string
		want want
	}{
		{
			name: "success",
			want: want{
				statusCode: http.StatusOK,
				resp:       string(successResp),
			},
		},
		{
			name: "internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "get balance: some error",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusOK:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetBalance(1).Return(1234.56, nil)
				srv.Storage.Transaction().(*mocks.MockTransactionStorage).EXPECT().FindWithdrawsForUser(1).Return([]*model.Transaction{
					{
						ID:          1,
						UserID:      1,
						OrderNumber: "9278923470",
						Sum:         111.11,
						Type:        model.TransactionTypeWithdrawal,
					}, {
						ID:          2,
						UserID:      1,
						OrderNumber: "12345678903",
						Sum:         22.22,
						Type:        model.TransactionTypeWithdrawal,
					},
				}, nil)

			case http.StatusInternalServerError:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetBalance(1).Return(0.0, errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodGet, "/balance", nil)
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			getBalanceHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			if test.want.statusCode == http.StatusOK {
				require.JSONEq(t, test.want.resp, string(resultBody))
			} else {
				require.Equal(t, test.want.resp, string(resultBody))
			}
		})
	}
}

func TestWithdrawHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)
	a := app.NewApp(srv)
	a.Session = model.Session{
		Token:  "token",
		UserID: 1,
	}

	cc := &middleware.Context{
		App: a,
	}

	type want struct {
		statusCode int
		resp       string
	}

	tests := []struct {
		name string
		req  string
		want want
	}{
		{
			name: "success",
			req:  `{"order":"346436439","sum":123.45}`,
			want: want{
				statusCode: http.StatusOK,
				resp:       "",
			},
		},
		{
			name: "low balance",
			req:  `{"order":"346436439","sum":123.45}`,
			want: want{
				statusCode: http.StatusPaymentRequired,
				resp:       "low balance",
			},
		},
		{
			name: "wrong number",
			req:  `{"order":"222222","sum":123.45}`,
			want: want{
				statusCode: http.StatusUnprocessableEntity,
				resp:       "wrong order number",
			},
		},
		{
			name: "internal error",
			req:  `{"order":"346436439","sum":123.45}`,
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "get balance: some error",
			},
		},
		{
			name: "wrong request",
			req:  `not expected body`,
			want: want{
				statusCode: http.StatusBadRequest,
				resp:       "code=400, message=Syntax error: offset=2, error=invalid character 'o' in literal null (expecting 'u'), internal=invalid character 'o' in literal null (expecting 'u')",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusOK:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetBalance(1).Return(200.00, nil)
				srv.Storage.Transaction().(*mocks.MockTransactionStorage).EXPECT().Create(gomock.Any()).Return(nil)
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().UpdateBalance(1, gomock.Any()).Return(nil)

			case http.StatusPaymentRequired:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetBalance(1).Return(100.00, nil)

			case http.StatusInternalServerError:
				srv.Storage.User().(*mocks.MockUserStorage).EXPECT().GetBalance(1).Return(0.0, errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodPost, "/user/balance/withdraw", strings.NewReader(test.req))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			withdrawHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			require.Equal(t, test.want.resp, string(resultBody))
		})
	}
}

func TestGetWithdrawalsHandler(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	srv := testServer(&config.Config{}, mockController)
	a := app.NewApp(srv)
	a.Session = model.Session{
		Token:  "token",
		UserID: 1,
	}

	cc := &middleware.Context{
		App: a,
	}

	someDate := time.Date(2024, 6, 30, 4, 52, 22, 0, time.UTC)

	withdrawals := []model.WithdrawalData{
		{
			OrderNumber: "9278923470",
			Sum:         111.11,
			ProcessedAt: &someDate,
		},
		{
			OrderNumber: "12345678903",
			Sum:         22.22,
			ProcessedAt: &someDate,
		},
	}

	successResp, err := json.Marshal(withdrawals)
	require.NoError(t, err)

	type want struct {
		statusCode int
		resp       string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "success",
			want: want{
				statusCode: http.StatusOK,
				resp:       string(successResp),
			},
		},
		{
			name: "no withdrawals",
			want: want{
				statusCode: http.StatusNoContent,
				resp:       "",
			},
		},
		{
			name: "internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
				resp:       "find withdrawals for user: some error",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.want.statusCode {
			case http.StatusOK:
				srv.Storage.Transaction().(*mocks.MockTransactionStorage).EXPECT().FindWithdrawsForUser(1).Return([]*model.Transaction{
					{
						ID:          1,
						UserID:      1,
						OrderNumber: "9278923470",
						Sum:         111.11,
						Type:        model.TransactionTypeWithdrawal,
						ProcessedAt: someDate,
					},
					{
						ID:          2,
						UserID:      1,
						OrderNumber: "12345678903",
						Sum:         22.22,
						Type:        model.TransactionTypeWithdrawal,
						ProcessedAt: someDate,
					},
				}, nil)

			case http.StatusNoContent:
				srv.Storage.Transaction().(*mocks.MockTransactionStorage).EXPECT().FindWithdrawsForUser(1).Return(nil, sql.ErrNoRows)

			case http.StatusInternalServerError:
				srv.Storage.Transaction().(*mocks.MockTransactionStorage).EXPECT().FindWithdrawsForUser(1).Return(nil, errors.New("some error"))
			}

			request := httptest.NewRequest(http.MethodGet, "/user/withdrawals", nil)
			recorder := httptest.NewRecorder()

			c := cc.App.Server.Echo.NewContext(request, recorder)
			cc.Context = c

			getWithdrawalsHandler(cc)

			result := recorder.Result()
			resultBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.statusCode, result.StatusCode)
			if test.want.statusCode == http.StatusOK {
				require.JSONEq(t, test.want.resp, string(resultBody))
			} else {
				require.Equal(t, test.want.resp, string(resultBody))
			}
		})
	}
}
