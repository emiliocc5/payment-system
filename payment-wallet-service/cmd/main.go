package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"time"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/metrics"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/pubsub/kafka"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/http"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/pubsub/rabbit"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/storage/postgresql"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/balance"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/payments"
	"github.com/emiliocc5/payment-system/payment-wallet-service/pkg/config"
	"github.com/emiliocc5/payment-system/payment-wallet-service/pkg/logger"
	"github.com/emiliocc5/payment-system/payment-wallet-service/pkg/signals"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_configPath = "./configuration"
)

func main() {
	configFile := flag.String("config-file", "config.yaml", "config file")
	flag.Parse()

	cfg, err := config.Parse(_configPath, *configFile)
	if err != nil {
		panic(err)
	}

	l := logger.New(cfg)

	run(l, cfg)
}

func run(logger *slog.Logger, cfg *config.Config) {
	ctx := context.Background()

	if err := migration(ctx, logger, cfg); err != nil {
		logger.Error("failed to run migrations", "error", err)
		panic(err)
	}

	srvCfg, asyncCfg, err := wire(ctx, logger, cfg)
	if err != nil {
		logger.Error("failed to wire services", "error", err)
		panic(err)
	}

	consumer, errNewConsumer := kafka.NewService(asyncCfg)
	if errNewConsumer != nil {
		logger.Error("failed to create kafka consumer", "error", errNewConsumer)
		panic(errNewConsumer)
	}
	if err := consumer.Start(); err != nil {
		logger.Error("Failed to start Kafka consumer", "error", err)
		return
	}

	srv := http.NewServer(srvCfg, logger)
	httpSrv, healthy := srv.ListenAndServe(ctx)

	// graceful shutdown
	stopCh := signals.SetupSignalHandler()
	sd, _ := signals.NewShutdown(3*time.Second, logger)
	sd.Graceful(stopCh, httpSrv, healthy)
}

func migration(ctx context.Context, logger *slog.Logger, cfg *config.Config) error {
	logger.Info("running database migrations")

	m, err := migrate.New(
		"file://migrations",
		cfg.StorageConfig.Dsn+"?sslmode=disable",
	)
	if err != nil {
		return err
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	logger.Info("migrations applied successfully")
	return nil
}

func wire(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*http.ServerConfig, *kafka.ServiceConfig, error) {
	var (
		balanceServiceConfig  balance.ServiceConfig
		paymentsServiceConfig payments.ServiceConfig
		pubConfig             rabbit.Config
		httpSrvCfg            http.ServerConfig
		consumerConfig        kafka.ConsumerConfig
		asyncSrvCfg           kafka.ServiceConfig
	)
	db, err := postgresql.NewDatabase(ctx, cfg.StorageConfig.Dsn)
	if err != nil {
		return nil, nil, err
	}

	balanceRepo := postgresql.NewPgBalanceRepository(db.DB)
	paymentRepo := postgresql.NewPgPaymentsRepository(db.DB)

	pubConfig.RoutingKey = cfg.PubConfig.RoutingKey
	pubConfig.Exchange = cfg.PubConfig.Exchange
	pubConfig.RabbitURL = cfg.PubConfig.RabbitURL
	pubConfig.Logger = logger

	pub, errRabbitPub := rabbit.NewRabbitPub(pubConfig)
	if errRabbitPub != nil {
		return nil, nil, errRabbitPub
	}

	prometheusMetrics := metrics.NewPrometheusMetrics()

	balanceServiceConfig.BalanceRepository = balanceRepo
	balanceServiceConfig.Logger = logger
	balanceSvc := balance.NewBalanceService(&balanceServiceConfig)

	paymentsServiceConfig.PaymentRepository = paymentRepo
	paymentsServiceConfig.Logger = logger
	paymentsServiceConfig.BalanceService = balanceSvc
	paymentsServiceConfig.DB = db
	paymentsServiceConfig.PublisherService = pub
	paymentsServiceConfig.MetricsService = prometheusMetrics
	paymentsSvc := payments.NewPaymentService(paymentsServiceConfig)

	httpSrvCfg.Port = cfg.Port
	httpSrvCfg.PaymentService = paymentsSvc
	httpSrvCfg.BalanceService = balanceSvc

	consumerConfig.GroupID = cfg.SubConfig.GroupID
	consumerConfig.Topics = cfg.SubConfig.Topics
	consumerConfig.StartOldest = cfg.SubConfig.StartOldest
	consumerConfig.Brokers = cfg.SubConfig.Brokers

	asyncSrvCfg.Logger = logger
	asyncSrvCfg.PaymentService = paymentsSvc
	asyncSrvCfg.ConsumerConfig = consumerConfig

	return &httpSrvCfg, &asyncSrvCfg, nil
}
