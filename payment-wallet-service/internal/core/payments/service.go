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
	Pending = "PENDING"
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
		exists, err := s.paymentRepo.CheckIdempotency(ctx, *tx, request.IdempotencyKey)
		if err != nil {
			s.logger.Error("failed to check idempotency",
				slog.Any("error", err),
				slog.String("idempotency_key", request.IdempotencyKey))

			return domain.ErrCheckIdempotency
		}

		if exists {
			//Publish error business metric here

			return nil
		}

		err = s.balanceService.ReserveFunds(ctx, *tx, request.UserID, request.Amount)
		if err != nil {
			//Publish error business metric here

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

			return domain.ErrCreatePayment
		}

		s.metricsService.RecordTransactionCompleted("Payment", true)

		paymentInitiatedEvent := &domain.PaymentInitiatedEvent{
			UserID:        payment.UserID,
			ClientNumber:  payment.ClientNumber,
			ServiceID:     payment.ServiceID,
			Amount:        payment.Amount,
			TransactionID: payment.ID,
		}

		s.publisherService.Publish(ctx, paymentInitiatedEvent)

		s.logger.Info("Payment created")
		return nil
	})
}

func (s *Service) Update(ctx context.Context, paymentID, status string) error {
	return nil
}
