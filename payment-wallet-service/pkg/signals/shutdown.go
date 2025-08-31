package signals

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

type Shutdown struct {
	logger                *slog.Logger
	serverShutdownTimeout time.Duration
}

func NewShutdown(serverShutdownTimeout time.Duration, logger *slog.Logger) (*Shutdown, error) {
	srv := &Shutdown{
		logger:                logger,
		serverShutdownTimeout: serverShutdownTimeout,
	}

	return srv, nil
}

func (s *Shutdown) Graceful(stopCh <-chan struct{}, httpServer *http.Server, healthy *int32) {
	ctx := context.Background()

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(ctx, s.serverShutdownTimeout)
	defer cancel()

	// all calls to /health will fail from now on
	atomic.StoreInt32(healthy, 0)

	s.logger.Info("shutting down", slog.Duration("timeout", s.serverShutdownTimeout))

	// stop OpenTelemetry tracer provider

	// determine if HTTP server was started
	if httpServer != nil {
		if err := httpServer.Shutdown(ctx); err != nil {
			s.logger.Warn("HTTP server shutdown failed", slog.Any("error", err))
		}
	}
}
