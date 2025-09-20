package ports

import (
	"context"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -destination=./mocks/balance_ports_mock.go -package=mocks -source=balance.go

type BalanceRepository interface {
	Get(ctx context.Context, userID string) (*domain.Balance, error)
	Reserve(ctx context.Context, tx pgx.Tx, userID string, amount int64) error
	Release(ctx context.Context, userID string, amount int64) error
	Confirm(ctx context.Context, userID string, amount int64) error
}

type BalanceService interface {
	ReserveFunds(ctx context.Context, tx pgx.Tx, userID string, amount int64) error
	ReleaseFunds(ctx context.Context, userID string, amount int64) error
	ConfirmReserve(ctx context.Context, userID string, amount int64) error
}
