package payments

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/golang/mock/gomock"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewPaymentService(t *testing.T) {
	logger := slog.Default()
	mockDB := mocks.NewMockDatabase(gomock.NewController(t))
	mockPaymentRepo := mocks.NewMockPaymentRepository(gomock.NewController(t))
	mockBalanceService := mocks.NewMockBalanceService(gomock.NewController(t))
	mockPublisher := mocks.NewMockPublisher(gomock.NewController(t))
	mockMetrics := mocks.NewMockMetrics(gomock.NewController(t))

	config := ServiceConfig{
		Logger:            logger,
		DB:                mockDB,
		PaymentRepository: mockPaymentRepo,
		BalanceService:    mockBalanceService,
		PublisherService:  mockPublisher,
		MetricsService:    mockMetrics,
	}

	service := NewPaymentService(config)

	assert.NotNil(t, service)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockPaymentRepo, service.paymentRepo)
	assert.Equal(t, mockBalanceService, service.balanceService)
	assert.Equal(t, mockPublisher, service.publisherService)
}

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDatabase(ctrl)
	mockPaymentRepo := mocks.NewMockPaymentRepository(ctrl)
	mockBalanceService := mocks.NewMockBalanceService(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockMetrics := mocks.NewMockMetrics(ctrl)

	service := &Service{
		logger:           slog.Default(),
		db:               mockDB,
		paymentRepo:      mockPaymentRepo,
		balanceService:   mockBalanceService,
		publisherService: mockPublisher,
		metricsService:   mockMetrics,
	}

	ctx := context.Background()
	request := domain.CreatePaymentRequest{
		IdempotencyKey: "test-key-123",
		UserID:         "user-123",
		Amount:         10050,
		ServiceID:      "service-1",
		ClientNumber:   "client-456",
	}

	t.Run("successful payment creation", func(t *testing.T) {
		var capturedTx pgx.Tx
		mockDB.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(*pgx.Tx) error) error {
				dummyTx := new(pgx.Tx)
				return fn(dummyTx)
			},
		).Times(1)

		mockPaymentRepo.EXPECT().
			CheckIdempotency(ctx, gomock.Any(), request.IdempotencyKey).
			Return(false, nil).Times(1)

		mockBalanceService.EXPECT().
			ReserveFunds(ctx, gomock.Any(), request.UserID, request.Amount).
			Return(nil).Times(1)

		mockPaymentRepo.EXPECT().
			Create(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, tx pgx.Tx, payment domain.Payment) error {
				capturedTx = tx
				assert.Nil(t, capturedTx)
				assert.NotEmpty(t, payment.ID)
				assert.Equal(t, request.IdempotencyKey, payment.IdempotencyKey)
				assert.Equal(t, request.UserID, payment.UserID)
				assert.Equal(t, request.Amount, payment.Amount)
				assert.Equal(t, Pending, payment.Status)
				assert.Equal(t, request.ServiceID, payment.ServiceID)
				assert.Equal(t, request.ClientNumber, payment.ClientNumber)
				assert.False(t, payment.CreatedAt.IsZero())
				assert.False(t, payment.UpdatedAt.IsZero())
				return nil
			}).Times(1)

		mockPublisher.EXPECT().
			Publish(ctx, gomock.Any()).Return(nil).Times(1)

		mockMetrics.EXPECT().RecordTransactionStarted(gomock.Any(), true).Times(1)
		mockMetrics.EXPECT().RecordTransactionProcessingTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

		err := service.Create(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("idempotency key already exists", func(t *testing.T) {
		mockDB.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(*pgx.Tx) error) error {
				dummyTx := new(pgx.Tx)
				return fn(dummyTx)
			},
		)

		mockPaymentRepo.EXPECT().
			CheckIdempotency(ctx, gomock.Any(), request.IdempotencyKey).
			Return(true, nil)

		mockBalanceService.EXPECT().ReserveFunds(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockPaymentRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockPublisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0)
		mockMetrics.EXPECT().RecordTransactionCompleted(gomock.Any(), gomock.Any()).Times(0)
		mockMetrics.EXPECT().RecordTransactionProcessingTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockMetrics.EXPECT().RecordTransactionIdempotent(gomock.Any()).Times(1)

		err := service.Create(ctx, request)
		assert.NoError(t, err)
	})

	t.Run("error checking idempotency", func(t *testing.T) {
		expectedError := errors.New("database error")

		mockDB.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(*pgx.Tx) error) error {
				dummyTx := new(pgx.Tx)
				return fn(dummyTx)
			},
		)
		mockMetrics.EXPECT().RecordTransactionStarted(gomock.Any(), false).Times(1)
		mockMetrics.EXPECT().RecordTransactionProcessingTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockMetrics.EXPECT().RecordTransactionIdempotent(gomock.Any()).Times(0)

		mockPaymentRepo.EXPECT().
			CheckIdempotency(ctx, gomock.Any(), request.IdempotencyKey).
			Return(false, expectedError)

		err := service.Create(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrCheckIdempotency, err)
	})

	t.Run("error reserving funds", func(t *testing.T) {
		expectedError := errors.New("insufficient balance")

		mockDB.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(*pgx.Tx) error) error {
				dummyTx := new(pgx.Tx)
				return fn(dummyTx)
			},
		)

		mockPaymentRepo.EXPECT().
			CheckIdempotency(ctx, gomock.Any(), request.IdempotencyKey).
			Return(false, nil)

		mockBalanceService.EXPECT().
			ReserveFunds(ctx, gomock.Any(), request.UserID, request.Amount).
			Return(expectedError)

		mockMetrics.EXPECT().RecordTransactionStarted(gomock.Any(), false).Times(1)
		mockMetrics.EXPECT().RecordTransactionProcessingTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockMetrics.EXPECT().RecordTransactionIdempotent(gomock.Any()).Times(0)

		err := service.Create(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("error creating payment", func(t *testing.T) {
		expectedError := errors.New("create payment error")

		mockDB.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(*pgx.Tx) error) error {
				dummyTx := new(pgx.Tx)
				return fn(dummyTx)
			},
		)

		mockPaymentRepo.EXPECT().
			CheckIdempotency(ctx, gomock.Any(), request.IdempotencyKey).
			Return(false, nil)

		mockBalanceService.EXPECT().
			ReserveFunds(ctx, gomock.Any(), request.UserID, request.Amount).
			Return(nil)

		mockPaymentRepo.EXPECT().
			Create(ctx, gomock.Any(), gomock.Any()).
			Return(expectedError)

		mockPublisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0)

		mockMetrics.EXPECT().RecordTransactionStarted(gomock.Any(), false).Times(1)
		mockMetrics.EXPECT().RecordTransactionProcessingTime(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockMetrics.EXPECT().RecordTransactionIdempotent(gomock.Any()).Times(0)

		err := service.Create(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrCreatePayment, err)
	})

	t.Run("transaction rollback on error", func(t *testing.T) {
		expectedError := errors.New("transaction error")

		mockDB.EXPECT().WithTx(ctx, gomock.Any()).Return(expectedError)

		err := service.Create(ctx, request)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func BenchmarkService_Create(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDatabase(ctrl)
	mockPaymentRepo := mocks.NewMockPaymentRepository(ctrl)
	mockBalanceService := mocks.NewMockBalanceService(ctrl)
	mockPublisher := mocks.NewMockPublisher(ctrl)
	mockMetrics := mocks.NewMockMetrics(ctrl)

	service := &Service{
		logger:           slog.Default(),
		db:               mockDB,
		paymentRepo:      mockPaymentRepo,
		balanceService:   mockBalanceService,
		publisherService: mockPublisher,
		metricsService:   mockMetrics,
	}

	ctx := context.Background()
	request := domain.CreatePaymentRequest{
		IdempotencyKey: "bench-key",
		UserID:         "user-123",
		Amount:         100.0,
		ServiceID:      "service-1",
		ClientNumber:   "client-456",
	}

	mockDB.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(*pgx.Tx) error) error {
			dummyTx := new(pgx.Tx)
			return fn(dummyTx)
		},
	).AnyTimes()

	mockPaymentRepo.EXPECT().CheckIdempotency(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	mockBalanceService.EXPECT().ReserveFunds(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockPaymentRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockPublisher.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordTransactionCompleted(gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordTransactionProcessingTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.Create(ctx, request)
	}
}

func TestService_Update(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		paymentID                  string
		status                     string
		getPaymentRepoResp         *domain.Payment
		getPaymentRepoErr          error
		confirmReserveServiceErr   error
		releaseReserveServiceErr   error
		updatePaymentRepoErr       error
		getPaymentServiceTimes     int
		confirmReserveServiceTimes int
		releaseReserveServiceTimes int
		updatePaymentRepoTimes     int
		successMetric              bool
		publishMetricTimes         int
		expectedError              error
	}{
		{
			name:      "Success update payment",
			paymentID: "payment-id",
			status:    Success,
			getPaymentRepoResp: &domain.Payment{
				ID:           "payment-id",
				Amount:       100.0,
				ServiceID:    "service-1",
				ClientNumber: "client-456",
				Status:       Pending,
			},
			getPaymentServiceTimes:     1,
			confirmReserveServiceTimes: 1,
			updatePaymentRepoTimes:     1,
			successMetric:              true,
			publishMetricTimes:         1,
		},
		{
			name:                       "Error getting payment",
			paymentID:                  "payment-id",
			status:                     Success,
			getPaymentRepoErr:          errors.New("get payment error"),
			getPaymentServiceTimes:     1,
			confirmReserveServiceTimes: 0,
			updatePaymentRepoTimes:     0,
			publishMetricTimes:         0,
			expectedError:              domain.ErrGetPayment,
		},
		{
			name:      "Payment update idempotency",
			paymentID: "payment-id",
			status:    Success,
			getPaymentRepoResp: &domain.Payment{
				ID:           "payment-id",
				Amount:       100.0,
				ServiceID:    "service-1",
				ClientNumber: "client-456",
				Status:       Success,
			},
			getPaymentServiceTimes:     1,
			confirmReserveServiceTimes: 0,
			updatePaymentRepoTimes:     0,
			publishMetricTimes:         0,
		},
		{
			name:      "Payment not success then release funds",
			paymentID: "payment-id",
			status:    "Error",
			getPaymentRepoResp: &domain.Payment{
				ID:           "payment-id",
				Amount:       100.0,
				ServiceID:    "service-1",
				ClientNumber: "client-456",
				Status:       Pending,
			},
			getPaymentServiceTimes:     1,
			releaseReserveServiceTimes: 1,
			updatePaymentRepoTimes:     1,
			successMetric:              true,
			publishMetricTimes:         1,
		},
		{
			name:      "Success update payment",
			paymentID: "payment-id",
			status:    Success,
			getPaymentRepoResp: &domain.Payment{
				ID:           "payment-id",
				Amount:       100.0,
				ServiceID:    "service-1",
				ClientNumber: "client-456",
				Status:       Pending,
			},
			getPaymentServiceTimes:     1,
			confirmReserveServiceTimes: 1,
			confirmReserveServiceErr:   errors.New("confirm reserve error"),
			updatePaymentRepoTimes:     0,
			publishMetricTimes:         0,
			expectedError:              domain.ErrConfirmReserve,
		},
		{
			name:      "Success update payment",
			paymentID: "payment-id",
			status:    Success,
			getPaymentRepoResp: &domain.Payment{
				ID:           "payment-id",
				Amount:       100.0,
				ServiceID:    "service-1",
				ClientNumber: "client-456",
				Status:       Pending,
			},
			getPaymentServiceTimes:     1,
			confirmReserveServiceTimes: 1,
			updatePaymentRepoTimes:     1,
			updatePaymentRepoErr:       errors.New("update payment error"),
			publishMetricTimes:         0,
			expectedError:              domain.ErrUpdatePayment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockPaymentRepo := mocks.NewMockPaymentRepository(ctrl)
			mockBalanceService := mocks.NewMockBalanceService(ctrl)
			mockMetrics := mocks.NewMockMetrics(ctrl)
			mockDB := mocks.NewMockDatabase(ctrl)

			mockDB.EXPECT().WithTx(context.Background(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(*pgx.Tx) error) error {
					dummyTx := new(pgx.Tx)
					return fn(dummyTx)
				},
			).AnyTimes()

			mockPaymentRepo.EXPECT().Get(context.Background(), tt.paymentID).
				Return(tt.getPaymentRepoResp, tt.getPaymentRepoErr).Times(tt.getPaymentServiceTimes)
			mockBalanceService.EXPECT().ConfirmReserve(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.confirmReserveServiceErr).Times(tt.confirmReserveServiceTimes)
			mockBalanceService.EXPECT().ReleaseFunds(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.releaseReserveServiceErr).Times(tt.releaseReserveServiceTimes)
			mockPaymentRepo.EXPECT().Update(context.Background(), gomock.Any(), gomock.Any()).
				Return(tt.updatePaymentRepoErr).Times(tt.updatePaymentRepoTimes)
			mockMetrics.EXPECT().RecordTransactionCompleted(PaymentTransactionType, tt.status).
				Times(tt.publishMetricTimes)
			mockMetrics.EXPECT().RecordTransactionProcessingTime(PaymentTransactionType, tt.status, gomock.Any()).
				Times(1)

			cfg := ServiceConfig{
				PaymentRepository: mockPaymentRepo,
				BalanceService:    mockBalanceService,
				MetricsService:    mockMetrics,
				DB:                mockDB,
				Logger:            slog.Default(),
			}

			paymentService := NewPaymentService(cfg)

			err := paymentService.Update(context.Background(), tt.paymentID, tt.status)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}
