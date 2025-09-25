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

	errReserve := s.balanceRepo.Reserve(ctx, tx, userID, amount)
	if errReserve != nil {
		s.logger.Error("failed to reserve funds",
			slog.Any("error", errReserve),
			slog.String("user_id", userID))

		return domain.ErrReserveFunds
	}

	return nil
}

func (s *Service) ReleaseFunds(ctx context.Context, tx pgx.Tx, userID string, amount int64) error {
	balance, errGetBalance := s.balanceRepo.Get(ctx, userID)
	if errGetBalance != nil {
		s.logger.Error("failed to get user balance",
			slog.Any("error", errGetBalance),
			slog.String("user_id", userID))

		return domain.ErrGetBalance
	}
	if balance.Reserved < amount {
		return domain.ErrInsufficientFunds
	}

	errRelease := s.balanceRepo.Release(ctx, tx, userID, amount)
	if errRelease != nil {
		s.logger.
			With("error", errRelease).
			Error("failed to release funds")

		return domain.ErrReleaseFunds
	}

	return nil
}

func (s *Service) ConfirmReserve(ctx context.Context, tx pgx.Tx, userID string, amount int64) error {
	balance, errGetBalance := s.balanceRepo.Get(ctx, userID)
	if errGetBalance != nil {
		s.logger.Error("failed to get user balance",
			slog.Any("error", errGetBalance),
			slog.String("user_id", userID))

		return domain.ErrGetBalance
	}
	if balance.Reserved < amount {
		return domain.ErrInsufficientFunds
	}

	errConfirm := s.balanceRepo.Confirm(ctx, tx, userID, amount)
	if errConfirm != nil {
		s.logger.
			With("error", errConfirm).
			Error("failed to confirm funds")

		return domain.ErrConfirmReserve
	}

	return nil
}
