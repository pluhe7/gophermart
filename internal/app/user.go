package app

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/pluhe7/gophermart/internal/model"
)

func (a *App) CreateUser(userData model.UserLoginData) error {
	user, err := a.Server.Storage.User().GetByLogin(userData.Login)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get user: %w", err)
	}

	if user != nil {
		return model.ErrUserExist
	}

	passwordHash, err := generatePasswordHash(userData.Password)
	if err != nil {
		return fmt.Errorf("bcrypt generage hash: %w", err)
	}

	err = a.Server.Storage.User().Create(&model.User{
		Login:        userData.Login,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (a *App) Login(loginData model.UserLoginData) (*model.Session, error) {
	user, err := a.Server.Storage.User().GetByLogin(loginData.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotExist
		}

		return nil, fmt.Errorf("get user: %w", err)
	}

	if !checkPassword(user.PasswordHash, loginData.Password) {
		return nil, model.ErrWrongPassword
	}

	session, err := a.CreateSession(user.ID)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

func (a *App) GetBalance() (*model.Balance, error) {
	balance, err := a.Server.Storage.User().GetBalance(a.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	withdraws, err := a.Server.Storage.Transaction().FindWithdrawsForUser(a.Session.UserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get withdrawns for user: %w", err)
	}

	var withdrawsAmount float64
	for _, withdraw := range withdraws {
		withdrawsAmount += withdraw.Sum
	}

	return &model.Balance{
		Current:   balance,
		Withdrawn: math.Round(withdrawsAmount*100) / 100,
	}, nil
}

func (a *App) Withdraw(withdrawal model.WithdrawalData) error {
	if !checkLuhn(withdrawal.OrderNumber) {
		return model.ErrWrongOrderNumber
	}

	balance, err := a.Server.Storage.User().GetBalance(a.Session.UserID)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	if withdrawal.Sum > balance {
		return model.ErrLowBalance
	}

	err = a.Server.Storage.Transaction().Create(&model.Transaction{
		UserID:      a.Session.UserID,
		OrderNumber: withdrawal.OrderNumber,
		Sum:         withdrawal.Sum,
		Type:        model.TransactionTypeWithdrawal,
	})
	if err != nil {
		return fmt.Errorf("create withdrawal transaction: %w", err)
	}

	err = a.Server.Storage.User().UpdateBalance(a.Session.UserID, -1*withdrawal.Sum)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}

func (a *App) GetWithdrawals() ([]model.WithdrawalData, error) {
	withdrawalTransactions, err := a.Server.Storage.Transaction().FindWithdrawsForUser(a.Session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoWithdrawals
		}

		return nil, fmt.Errorf("find withdrawals for user: %w", err)
	}

	withdrawals := make([]model.WithdrawalData, 0, len(withdrawalTransactions))
	for _, transaction := range withdrawalTransactions {
		withdrawals = append(withdrawals, model.WithdrawalData{
			OrderNumber: transaction.OrderNumber,
			Sum:         transaction.Sum,
			ProcessedAt: &transaction.ProcessedAt,
		})
	}

	return withdrawals, nil
}
