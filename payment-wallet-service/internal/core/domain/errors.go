package domain

import "errors"

var (
	ErrGetBalance        = errors.New("failed to get user balance")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrReserveFunds      = errors.New("failed to reserve funds")
	ErrCreatePayment     = errors.New("failed to create payment")
	ErrCheckIdempotency  = errors.New("failed to check idempotency")
	ErrGetPayment        = errors.New("failed to get payment")
	ErrConfirmReserve    = errors.New("failed to confirm reserve")
	ErrReleaseFunds      = errors.New("failed to release funds")
	ErrUpdatePayment     = errors.New("failed to update payment")
)
