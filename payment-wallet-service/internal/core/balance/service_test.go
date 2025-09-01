package balance

import (
	"context"
	"errors"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"log/slog"
	"testing"
	"time"
)

func TestNewBalanceService(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockBalanceRepository(ctrl)
	logger := slog.Default()

	config := &ServiceConfig{
		Logger:            logger,
		BalanceRepository: mockRepo,
	}

	service := NewBalanceService(config)

	assert.NotNil(t, service)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, mockRepo, service.balanceRepo)
}

func TestService_ReserveFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBalanceRepo := mocks.NewMockBalanceRepository(ctrl)
	logger := slog.Default()

	service := &Service{
		logger:      logger,
		balanceRepo: mockBalanceRepo,
	}

	ctx := context.Background()
	tx := new(pgx.Tx)
	userID := "valid-user-id"
	amount := int64(10)

	t.Run("successful reserve", func(t *testing.T) {
		mockBalanceRepo.EXPECT().Get(ctx, userID).Return(&domain.Balance{
			UserID:    userID,
			Available: amount,
			Reserved:  0,
			UpdatedAt: time.Time{},
		}, nil).Times(1)

		mockBalanceRepo.EXPECT().
			ReserveFunds(ctx, gomock.Any(), userID, amount).Return(nil).Times(1)

		err := service.ReserveFunds(ctx, *tx, userID, amount)
		assert.NoError(t, err)
	})

	t.Run("failed to get user balance", func(t *testing.T) {
		mockBalanceRepo.EXPECT().Get(ctx, userID).Return(nil, pgx.ErrNoRows)
		err := service.ReserveFunds(ctx, *tx, userID, amount)
		assert.Error(t, err)
		assert.Equal(t, err, domain.ErrGetBalance)
	})

	t.Run("insufficient funds - amount exceeds available", func(t *testing.T) {
		mockBalanceRepo.EXPECT().Get(ctx, userID).Return(&domain.Balance{
			UserID:    userID,
			Available: 5,
			Reserved:  0,
			UpdatedAt: time.Time{},
		}, nil).Times(1)

		err := service.ReserveFunds(ctx, *tx, userID, amount)
		assert.Error(t, err)
		assert.Equal(t, err, domain.ErrInsufficientFunds)
	})

	t.Run("failed to reserve funds in repository", func(t *testing.T) {
		mockBalanceRepo.EXPECT().Get(ctx, userID).Return(&domain.Balance{
			UserID:    userID,
			Available: amount,
			Reserved:  0,
			UpdatedAt: time.Time{},
		}, nil).Times(1)
		mockBalanceRepo.EXPECT().ReserveFunds(ctx, gomock.Any(), userID, amount).
			Return(errors.New("error reserving funds")).Times(1)

		err := service.ReserveFunds(ctx, *tx, userID, amount)
		assert.Error(t, err)
		assert.Equal(t, err, domain.ErrReserveFunds)
	})
}
