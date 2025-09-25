package rabbit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/emiliocc5/payment-system/payment-wallet-service/pkg/uidgen"
	"github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Logger     *slog.Logger
	RabbitURL  string
	Exchange   string
	RoutingKey string
}

type Pub struct {
	logger     *slog.Logger
	conn       *amqp091.Connection
	channel    *amqp091.Channel
	exchange   string
	routingKey string
}

func NewRabbitPub(config Config) (*Pub, error) {
	conn, err := amqp091.Dial(config.RabbitURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	err = channel.ExchangeDeclare(
		config.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, err
	}

	return &Pub{
		logger:     config.Logger,
		conn:       conn,
		channel:    channel,
		exchange:   config.Exchange,
		routingKey: config.RoutingKey,
	}, nil
}

func (p *Pub) Publish(ctx context.Context, event *domain.PaymentInitiatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event", "error", err)
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
			MessageId:   uidgen.NewUUID(),
		},
	)

	if err != nil {
		p.logger.Error("failed to publish message", "error", err)

		return err
	}

	p.logger.Info("message published successfully", "routing_key", p.routingKey)
	return nil
}

func (p *Pub) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
