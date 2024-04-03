package app

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/logger"
	"github.com/pluhe7/gophermart/internal/model"
)

func (a *App) SendOrder(number string) error {
	if !checkLuhn(number) {
		return model.ErrOrderNumberFormat
	}

	existingOrder, err := a.Server.Storage.Order().Get(number)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get existing order: %w", err)
	}

	if existingOrder != nil {
		if existingOrder.UserID == a.Session.UserID {
			return model.ErrCurrentUserOrderExist
		}
		return model.ErrOtherUserOrderExist
	}

	err = a.Server.Storage.Order().Create(&model.Order{
		Number: number,
		UserID: a.Session.UserID,
	})
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

func checkLuhn(numberStr string) bool {
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		logger.Log.Error("check luhn: parse number", zap.Error(err))
		return false
	}

	if number <= 0 {
		return false
	}

	digits := splitNumber(number)
	sum := 0
	parity := len(digits) % 2

	for i := 0; i < len(digits); i++ {
		digit := digits[i]
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}

func splitNumber(number int) []int {
	var digits []int
	for number > 0 {
		digits = append(digits, number%10)
		number /= 10
	}

	var result []int
	for i := len(digits) - 1; i >= 0; i-- {
		result = append(result, digits[i])
	}

	return result
}

func (a *App) GetOrders() ([]model.Order, error) {
	orders, err := a.Server.Storage.Order().FindForUser(a.Session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNoOrders
		}

		return nil, fmt.Errorf("find orders for user: %w", err)
	}

	return orders, nil
}
