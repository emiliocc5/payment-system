package balance

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
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
			Reserve(ctx, gomock.Any(), userID, amount).Return(nil).Times(1)

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
		mockBalanceRepo.EXPECT().Reserve(ctx, gomock.Any(), userID, amount).
			Return(errors.New("error reserving funds")).Times(1)

		err := service.ReserveFunds(ctx, *tx, userID, amount)
		assert.Error(t, err)
		assert.Equal(t, err, domain.ErrReserveFunds)
	})
}

func TestService_ReleaseFunds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		userID                  string
		amount                  int64
		getBalanceResponse      *domain.Balance
		getBalanceError         error
		releaseBalanceRepoError error
		getBalanceTimes         int
		releaseBalanceTimes     int
		expectedError           error
	}{
		{
			name:   "successful balance release",
			userID: "valid-user-id",
			amount: 10,
			getBalanceResponse: &domain.Balance{
				UserID:    "valid-user-id",
				Available: 0,
				Reserved:  10,
			},
			getBalanceTimes:     1,
			releaseBalanceTimes: 1,
		},
		{
			name:                "error getting balance",
			userID:              "valid-user-id",
			amount:              10,
			getBalanceTimes:     1,
			releaseBalanceTimes: 0,
			getBalanceError:     errors.New("error getting balance"),
			expectedError:       domain.ErrGetBalance,
		},
		{
			name:   "insufficient balance to release",
			userID: "valid-user-id",
			amount: 10,
			getBalanceResponse: &domain.Balance{
				UserID:    "valid-user-id",
				Available: 0,
				Reserved:  9,
			},
			getBalanceTimes:     1,
			releaseBalanceTimes: 0,
			expectedError:       domain.ErrInsufficientFunds,
		},
		{
			name:   "fail releasing balance",
			userID: "valid-user-id",
			amount: 10,
			getBalanceResponse: &domain.Balance{
				UserID:    "valid-user-id",
				Available: 0,
				Reserved:  10,
			},
			getBalanceTimes:         1,
			releaseBalanceTimes:     1,
			releaseBalanceRepoError: errors.New("error releasing funds"),
			expectedError:           domain.ErrReleaseFunds,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockBalanceRepository(ctrl)

			mockRepo.EXPECT().Get(context.Background(), tt.userID).
				Return(tt.getBalanceResponse, tt.getBalanceError).Times(tt.getBalanceTimes)

			mockRepo.EXPECT().Release(context.Background(), tt.userID, tt.amount).
				Return(tt.releaseBalanceRepoError).Times(tt.releaseBalanceTimes)

			cfg := &ServiceConfig{
				Logger:            slog.Default(),
				BalanceRepository: mockRepo,
			}
			service := NewBalanceService(cfg)

			errReleaseFunds := service.ReleaseFunds(context.Background(), tt.userID, tt.amount)

			assert.Equal(t, tt.expectedError, errReleaseFunds)
		})
	}
}

func TestService_ConfirmReserve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		userID                  string
		amount                  int64
		getBalanceResponse      *domain.Balance
		getBalanceError         error
		confirmBalanceRepoError error
		getBalanceTimes         int
		confirmBalanceTimes     int
		expectedError           error
	}{
		{
			name:   "successful balance confirm",
			userID: "valid-user-id",
			amount: 10,
			getBalanceResponse: &domain.Balance{
				UserID:    "valid-user-id",
				Available: 0,
				Reserved:  10,
			},
			getBalanceTimes:     1,
			confirmBalanceTimes: 1,
		},
		{
			name:                "error getting balance",
			userID:              "valid-user-id",
			amount:              10,
			getBalanceTimes:     1,
			confirmBalanceTimes: 0,
			getBalanceError:     errors.New("error getting balance"),
			expectedError:       domain.ErrGetBalance,
		},
		{
			name:   "insufficient balance to confirm",
			userID: "valid-user-id",
			amount: 10,
			getBalanceResponse: &domain.Balance{
				UserID:    "valid-user-id",
				Available: 0,
				Reserved:  9,
			},
			getBalanceTimes:     1,
			confirmBalanceTimes: 0,
			expectedError:       domain.ErrInsufficientFunds,
		},
		{
			name:   "fail confirming balance",
			userID: "valid-user-id",
			amount: 10,
			getBalanceResponse: &domain.Balance{
				UserID:    "valid-user-id",
				Available: 0,
				Reserved:  10,
			},
			getBalanceTimes:         1,
			confirmBalanceTimes:     1,
			confirmBalanceRepoError: errors.New("error confirming"),
			expectedError:           domain.ErrConfirmReserve,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRepo := mocks.NewMockBalanceRepository(ctrl)

			mockRepo.EXPECT().Get(context.Background(), tt.userID).
				Return(tt.getBalanceResponse, tt.getBalanceError).Times(tt.getBalanceTimes)

			mockRepo.EXPECT().Confirm(context.Background(), tt.userID, tt.amount).
				Return(tt.confirmBalanceRepoError).Times(tt.confirmBalanceTimes)

			cfg := &ServiceConfig{
				Logger:            slog.Default(),
				BalanceRepository: mockRepo,
			}
			service := NewBalanceService(cfg)

			errReleaseFunds := service.ConfirmReserve(context.Background(), tt.userID, tt.amount)

			assert.Equal(t, tt.expectedError, errReleaseFunds)
		})
	}
}
