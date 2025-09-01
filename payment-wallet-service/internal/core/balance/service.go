package balance

import (
	"context"
	"log/slog"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports"
	"github.com/jackc/pgx/v5"
)

type ServiceConfig struct {
	Logger            *slog.Logger
	BalanceRepository ports.BalanceRepository
}

type Service struct {
	logger      *slog.Logger
	balanceRepo ports.BalanceRepository
}

func NewBalanceService(config *ServiceConfig) *Service {
	return &Service{
		logger:      config.Logger,
		balanceRepo: config.BalanceRepository,
	}
}

func (s *Service) ReserveFunds(ctx context.Context, tx pgx.Tx, userID string, amount int64) error {
	balance, errGetBalance := s.balanceRepo.Get(ctx, userID)
	if errGetBalance != nil {
		s.logger.Error("failed to get user balance",
			slog.Any("error", errGetBalance),
			slog.String("user_id", userID))

		return domain.ErrGetBalance
	}

	if balance.Available < amount {
		return domain.ErrInsufficientFunds
	}

	errReserve := s.balanceRepo.ReserveFunds(ctx, tx, userID, amount)
	if errReserve != nil {
		s.logger.Error("failed to reserve funds",
			slog.Any("error", errReserve),
			slog.String("user_id", userID))

		return domain.ErrReserveFunds
	}

	return nil
}

func (s *Service) Update(ctx context.Context, userID string, amount int64) error {
	return nil
}
