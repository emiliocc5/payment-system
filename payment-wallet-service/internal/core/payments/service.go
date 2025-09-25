package payments

import (
	"context"
	"log/slog"
	"time"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports"
	"github.com/emiliocc5/payment-system/payment-wallet-service/pkg/uidgen"
	"github.com/jackc/pgx/v5"
)

const (
	Pending                = "PENDING"
	Success                = "SUCCESS"
	PaymentTransactionType = "Payment"
)

type ServiceConfig struct {
	Logger            *slog.Logger
	DB                ports.Database
	PaymentRepository ports.PaymentRepository
	BalanceService    ports.BalanceService
	PublisherService  ports.Publisher
	MetricsService    ports.Metrics
}

type Service struct {
	logger           *slog.Logger
	db               ports.Database
	paymentRepo      ports.PaymentRepository
	balanceService   ports.BalanceService
	publisherService ports.Publisher
	metricsService   ports.Metrics
}

func NewPaymentService(config ServiceConfig) *Service {
	return &Service{
		logger:           config.Logger,
		paymentRepo:      config.PaymentRepository,
		balanceService:   config.BalanceService,
		db:               config.DB,
		publisherService: config.PublisherService,
		metricsService:   config.MetricsService,
	}
}

func (s *Service) Create(ctx context.Context, request domain.CreatePaymentRequest) error {
	return s.db.WithTx(ctx, func(tx *pgx.Tx) error {
		start := time.Now()

		defer func() {
			duration := time.Since(start)
			s.metricsService.RecordTransactionProcessingTime(PaymentTransactionType, Pending, duration)
		}()

		exists, err := s.paymentRepo.CheckIdempotency(ctx, *tx, request.IdempotencyKey)
		if err != nil {
			s.logger.Error("failed to check idempotency",
				slog.Any("error", err),
				slog.String("idempotency_key", request.IdempotencyKey))

			s.metricsService.RecordTransactionStarted(PaymentTransactionType, false)

			return domain.ErrCheckIdempotency
		}

		if exists {
			s.metricsService.RecordTransactionIdempotent(PaymentTransactionType)

			return nil
		}

		err = s.balanceService.ReserveFunds(ctx, *tx, request.UserID, request.Amount)
		if err != nil {
			s.metricsService.RecordTransactionStarted(PaymentTransactionType, false)

			return err
		}

		payment := &domain.Payment{
			ID:             uidgen.NewUUID(),
			IdempotencyKey: request.IdempotencyKey,
			UserID:         request.UserID,
			Amount:         request.Amount,
			Status:         Pending,
			ServiceID:      request.ServiceID,
			ClientNumber:   request.ClientNumber,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		errCreate := s.paymentRepo.Create(ctx, *tx, *payment)
		if errCreate != nil {
			slog.Error("failed to create payment",
				slog.Any("error", errCreate),
				slog.String("user_id", request.UserID))
			s.metricsService.RecordTransactionStarted(PaymentTransactionType, false)

			return domain.ErrCreatePayment
		}

		s.metricsService.RecordTransactionStarted(PaymentTransactionType, true)

		paymentInitiatedEvent := &domain.PaymentInitiatedEvent{
			UserID:        payment.UserID,
			ClientNumber:  payment.ClientNumber,
			ServiceID:     payment.ServiceID,
			Amount:        payment.Amount,
			TransactionID: payment.ID,
		}

		errPublishPayment := s.publisherService.Publish(ctx, paymentInitiatedEvent)
		if errPublishPayment != nil {
			slog.Error("failed to create payment",
				slog.Any("error", errPublishPayment),
				slog.String("user_id", request.UserID))

			return errPublishPayment
		}

		s.logger.Debug("Payment created")
		return nil
	})
}

func (s *Service) Update(ctx context.Context, paymentID, status string) error {
	start := time.Now()
	payment, errGetPayment := s.paymentRepo.Get(ctx, paymentID)
	if errGetPayment != nil {
		s.logger.
			With("Error", errGetPayment).
			Error("failed to get payment")

		return domain.ErrGetPayment
	}

	if payment.Status == Success {
		s.logger.
			With("PaymentID", paymentID).
			Warn("Payment already processed")

		return nil
	}

	return s.db.WithTx(ctx, func(tx *pgx.Tx) error {
		defer func() {
			duration := time.Since(start)
			s.metricsService.RecordTransactionProcessingTime(PaymentTransactionType, status, duration)
		}()
		if status == Success {
			errConfirmReserve := s.balanceService.ConfirmReserve(ctx, *tx, payment.UserID, payment.Amount)
			if errConfirmReserve != nil {
				s.logger.
					With("Error", errConfirmReserve).
					Error("failed to confirm reserve")

				return domain.ErrConfirmReserve
			}
		} else {
			errReleaseFunds := s.balanceService.ReleaseFunds(ctx, *tx, payment.UserID, payment.Amount)
			if errReleaseFunds != nil {
				s.logger.
					With("Error", errReleaseFunds).
					Error("failed to release reserve")

				return domain.ErrReleaseFunds
			}
		}

		payment.Status = status

		errUpdatePayment := s.paymentRepo.Update(ctx, *tx, *payment)
		if errUpdatePayment != nil {
			s.logger.
				With("Error", errUpdatePayment).
				Error("failed to update payment")

			return domain.ErrUpdatePayment
		}

		s.logger.Debug("Payment updated")
		s.metricsService.RecordTransactionCompleted(PaymentTransactionType, status)

		return nil
	})
}
