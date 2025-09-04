package domain

import "errors"

var (
	ErrGetBalance        = errors.New("failed to get user balance")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrReserveFunds      = errors.New("failed to reserve funds")
	ErrCreatePayment     = errors.New("failed to create payment")
	ErrCheckIdempotency  = errors.New("failed to check idempotency")
)
