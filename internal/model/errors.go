package model

import "errors"

var (
	ErrUserNotExist  = errors.New("user not exist")
	ErrUserExist     = errors.New("user already exist")
	ErrWrongPassword = errors.New("wrong password")

	ErrLowBalance       = errors.New("low balance")
	ErrWrongOrderNumber = errors.New("wrong order number")
	ErrNoWithdrawals    = errors.New("withdrawals not found")

	ErrCurrentUserOrderExist = errors.New("order already was sent")
	ErrOtherUserOrderExist   = errors.New("other user already sent this order")
	ErrOrderNumberFormat     = errors.New("wrong order number")

	ErrNoOrders = errors.New("orders not found")
)
