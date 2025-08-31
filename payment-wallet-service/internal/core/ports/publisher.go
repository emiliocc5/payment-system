package ports

import (
	"context"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
)

type Publisher interface {
	Publish(ctx context.Context, event *domain.PaymentInitiatedEvent) error
}
