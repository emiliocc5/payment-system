package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/IBM/sarama"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_HandlePaymentEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		message                 *sarama.ConsumerMessage
		transactionID           string
		status                  string
		mockPaymentServiceError error
		mockPaymentServiceTimes int
		expectedError           error
	}{
		{
			name: "success",
			message: &sarama.ConsumerMessage{
				Key:   []byte("userID"),
				Value: getMockedMessage(),
			},
			transactionID:           "transaction_id",
			status:                  "success",
			mockPaymentServiceTimes: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockPaymentService := mocks.NewMockPaymentService(ctrl)
			mockPaymentService.EXPECT().Update(context.Background(), tt.transactionID, tt.status).
				Return(tt.mockPaymentServiceError).
				Times(tt.mockPaymentServiceTimes)

			serv := Service{
				logger:         slog.Default(),
				paymentService: mockPaymentService,
			}

			err := serv.handlePaymentEvent(context.Background(), tt.message)

			assert.Equal(t, tt.expectedError, err)
		})
	}

}

func getMockedMessage() []byte {
	var msg map[string]interface{}

	msg = map[string]interface{}{
		"transaction_id": "transaction_id",
		"status":         "success",
		"metadata":       map[string]interface{}{},
	}

	bytes, _ := json.Marshal(msg)
	return bytes
}
