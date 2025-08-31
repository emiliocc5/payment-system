package logger

import (
	"log/slog"

	"github.com/emiliocc5/payment-system/payment-wallet-service/pkg/config"
)

func New(cfg *config.Config) *slog.Logger {
	logger := slog.Default()

	logger.With(slog.String("service.name", cfg.AppID)).
		With(slog.String("service.version", cfg.Version))

	return logger
}
