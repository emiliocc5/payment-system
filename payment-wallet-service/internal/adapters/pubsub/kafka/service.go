package kafka

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports"
)

var (
	_healthy int32
)

type ConsumerConfig struct {
	Brokers     []string
	GroupID     string
	Topics      []string
	StartOldest bool
}

type ServiceConfig struct {
	Logger         *slog.Logger
	PaymentService ports.PaymentService
	Consumer       *sarama.ConsumerGroup
	ConsumerConfig ConsumerConfig
}

type Service struct {
	logger         *slog.Logger
	paymentService ports.PaymentService
	consumer       sarama.ConsumerGroup
	wg             *sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	topics         []string
	ready          chan bool
}

func NewService(config ServiceConfig) (*Service, error) {
	cfg := sarama.NewConfig()
	strategies := make([]sarama.BalanceStrategy, 0)
	strategies = append(strategies, sarama.NewBalanceStrategyRoundRobin())
	cfg.Consumer.Group.Rebalance.GroupStrategies = strategies
	cfg.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	cfg.Consumer.Group.Session.Timeout = 30 * time.Second

	if config.ConsumerConfig.StartOldest {
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	consumerGroup, err := sarama.NewConsumerGroup(config.ConsumerConfig.Brokers, config.ConsumerConfig.GroupID, cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		logger:         config.Logger,
		paymentService: config.PaymentService,
		consumer:       consumerGroup,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

func (s *Service) Start() error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			if err := s.consumer.Consume(s.ctx, s.topics, s); err != nil {
				s.logger.Error("Error from consumer", slog.Any("error", err))
				return
			}
			if s.ctx.Err() != nil {
				s.logger.Info("Consumer context cancelled")
				return
			}
			s.ready = make(chan bool)
		}
	}()

	<-s.ready
	s.logger.Info("Kafka consumer up and running",
		slog.Any("topics", s.topics))

	return nil
}

func (s *Service) Stop() error {
	s.logger.Info("Shutting down Kafka consumer")
	s.cancel()
	s.wg.Wait()

	if err := s.consumer.Close(); err != nil {
		s.logger.Error("Error closing consumer group", slog.Any("error", err))
		return err
	}

	s.logger.Info("Kafka consumer stopped")
	return nil
}

func (s *Service) Setup(_ sarama.ConsumerGroupSession) error {
	close(s.ready)
	return nil
}

func (s *Service) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Service) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			s.logger.Info("Received message",
				slog.String("topic", message.Topic),
				slog.Int64("partition", int64(message.Partition)),
				slog.Int64("offset", message.Offset))

			if err := s.handleMessage(message); err != nil {
				s.logger.Error("Error processing message",
					slog.Any("error", err),
					slog.String("topic", message.Topic),
					slog.String("key", string(message.Key)))

				//TODO decide if mark as processed or not
			}

			session.MarkMessage(message, "")

		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *Service) handleMessage(message *sarama.ConsumerMessage) error {
	switch message.Topic {
	case "payment-events":
		return s.handlePaymentEvent(message)
	default:
		s.logger.Warn("Unknown topic", slog.String("topic", message.Topic))
		return nil
	}
}
