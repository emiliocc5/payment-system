package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"time"

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

	srvCfg, err := wire(ctx, logger, cfg)
	if err != nil {
		logger.Error("failed to wire services", "error", err)
		panic(err)
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

func wire(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*http.ServerConfig, error) {
	var (
		balanceServiceConfig  balance.ServiceConfig
		paymentsServiceConfig payments.ServiceConfig
		pubConfig             rabbit.Config
		srvCfg                http.ServerConfig
	)
	db, err := postgresql.NewDatabase(ctx, cfg.StorageConfig.Dsn)
	if err != nil {
		return nil, err
	}

	balanceRepo := postgresql.NewPgBalanceRepository(db.DB)
	paymentRepo := postgresql.NewPgPaymentsRepository(db.DB)

	pubConfig.RoutingKey = cfg.PubConfig.RoutingKey
	pubConfig.Exchange = cfg.PubConfig.Exchange
	pubConfig.RabbitURL = cfg.PubConfig.RabbitURL
	pubConfig.Logger = logger

	pub, errRabbitPub := rabbit.NewRabbitPub(pubConfig)
	if errRabbitPub != nil {
		return nil, errRabbitPub
	}

	balanceServiceConfig.BalanceRepository = balanceRepo
	balanceServiceConfig.Logger = logger
	balanceSvc := balance.NewBalanceService(&balanceServiceConfig)

	paymentsServiceConfig.PaymentRepository = paymentRepo
	paymentsServiceConfig.Logger = logger
	paymentsServiceConfig.BalanceService = balanceSvc
	paymentsServiceConfig.DB = db
	paymentsServiceConfig.PublisherService = pub
	paymentsSvc := payments.NewPaymentService(paymentsServiceConfig)

	srvCfg.Port = cfg.Port
	srvCfg.PaymentService = paymentsSvc
	srvCfg.BalanceService = balanceSvc

	return &srvCfg, nil
}
