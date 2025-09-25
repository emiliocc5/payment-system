package storage

import "errors"

var (
	ErrInsufficientFunds         = errors.New("insufficient funds")
	ErrInsufficientReservedFunds = errors.New("insufficient reserved funds")
	ErrPaymentNotFound           = errors.New("payment not found")
)
