package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type Config struct {
	Brokers     []string
	GroupID     string
	Topics      []string
	Logger      *slog.Logger
	StartOldest bool
}
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topics        []string
	ready         chan bool
	logger        *slog.Logger
	handlers      map[string]MessageHandler
	mu            sync.RWMutex
}

type MessageHandler func(ctx context.Context, message *sarama.ConsumerMessage) error

func NewConsumer(config Config) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	rebalanceStrategies := append(make([]sarama.BalanceStrategy, 0), sarama.NewBalanceStrategyRoundRobin())
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = rebalanceStrategies
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	if config.StartOldest {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	saramaConfig.Consumer.MaxProcessingTime = 10 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topics:        config.Topics,
		ready:         make(chan bool),
		logger:        config.Logger,
		handlers:      make(map[string]MessageHandler),
	}, nil
}

func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[topic] = handler
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			c.logger.Info("Message received",
				slog.String("topic", message.Topic),
				slog.Int("partition", int(message.Partition)),
				slog.Int64("offset", message.Offset),
				slog.String("key", string(message.Key)),
			)

			c.mu.RLock()
			handler, exists := c.handlers[message.Topic]
			c.mu.RUnlock()

			if !exists {
				c.logger.Warn("No handler registered for topic", slog.String("topic", message.Topic))
				session.MarkMessage(message, "")
				continue
			}

			// Ejecutar handler
			ctx := context.Background()
			if err := handler(ctx, message); err != nil {
				c.logger.Error("Error processing message",
					slog.String("topic", message.Topic),
					slog.String("error", err.Error()),
				)
				// En caso de error, podrías decidir si hacer commit o no
				// Por ahora, marcamos el mensaje como procesado
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Debug("Starting Kafka consumer", slog.String("topics", c.topics))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				c.logger.Debug("Consumer context cancelled")
				return
			default:
				if err := c.consumerGroup.Consume(ctx, c.topics, c); err != nil {
					c.logger.Error("Error consuming messages", slog.String("error", err.Error()))
					return
				}
				if ctx.Err() != nil {
					return
				}
				c.ready = make(chan bool)
			}
		}
	}()

	<-c.ready
	c.logger.Debug("Kafka consumer ready")

	wg.Wait()
	return nil
}

func (c *Consumer) Close() error {
	c.logger.Info("Closing Kafka consumer")
	return c.consumerGroup.Close()
}

func (c *Consumer) IsReady() bool {
	select {
	case <-c.ready:
		return true
	default:
		return false
	}
}

// HandlePaymentMessage maneja mensajes de pagos
func HandlePaymentMessage(logger *slog.Logger) MessageHandler {
	return func(ctx context.Context, message *sarama.ConsumerMessage) error {
		var paymentMsg domain.PaymentResultEvent
		if err := json.Unmarshal(message.Value, &paymentMsg); err != nil {
			logger.Error("Failed to unmarshal payment message", slog.String("error", err.Error()))
			return err
		}

		logger.Info("Processing payment message",
			slog.String("payment_id", paymentMsg.ID),
			slog.String("user_id", paymentMsg.UserID),
			slog.Float64("amount", paymentMsg.Amount),
			slog.String("type", paymentMsg.Type),
		)

		// Aquí implementarías la lógica de negocio para procesar el pago
		// Por ejemplo: actualizar balance, crear transacción, etc.

		return nil
	}
}
