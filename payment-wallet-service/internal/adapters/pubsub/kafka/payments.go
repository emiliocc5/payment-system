package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
)

func (s *Service) handlePaymentEvent(ctx context.Context, message *sarama.ConsumerMessage) error {
	var event domain.PaymentResultEvent
	err := json.Unmarshal(message.Value, &event)
	if err != nil {
		s.logger.
			With("error", err.Error()).
			Error("error unmarshalling event")

		return err
	}
	event.UserID = string(message.Key)

	errUpdate := s.paymentService.Update(ctx, event.TransactionID, event.Status)
	if errUpdate != nil {
		s.logger.
			With("error", errUpdate.Error()).
			Error("error updating payment status")

		return errUpdate
	}

	return nil
}
