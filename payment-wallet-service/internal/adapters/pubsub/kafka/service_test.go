package kafka

import (
	"log/slog"
	"testing"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_ServiceStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	paymentMock := mocks.NewMockPaymentService(ctrl)

	config := ServiceConfig{
		Logger:         slog.Default(),
		PaymentService: paymentMock,
		ConsumerConfig: ConsumerConfig{
			Brokers:     []string{"localhost:29092"},
			GroupID:     "payment-events-consumer",
			Topics:      []string{"payment-events"},
			StartOldest: false,
		},
	}

	consumerService, errNewServ := NewService(&config)
	if errNewServ != nil {
		t.Fatal(errNewServ)
	}

	err := consumerService.Start()

	assert.Nil(t, err)
}
