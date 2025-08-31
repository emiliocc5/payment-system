package ports

import (
	"context"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/jackc/pgx/v5"
)

type PaymentRepository interface {
	CheckIdempotency(ctx context.Context, tx pgx.Tx, idempotencyKey string) (bool, error)
	Create(ctx context.Context, tx pgx.Tx, payment domain.Payment) error
	Update(ctx context.Context, payment domain.Payment) error
}

type PaymentService interface {
	Create(ctx context.Context, request domain.CreatePaymentRequest) error
	Update(ctx context.Context, paymentID, status string) error
}
