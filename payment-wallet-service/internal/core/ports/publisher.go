package ports

import (
	"context"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
)

//go:generate mockgen -destination=../mocks/publisher_ports_mock.go -package=mocks -source=publisher.go

type Publisher interface {
	Publish(ctx context.Context, event *domain.PaymentInitiatedEvent) error
}
