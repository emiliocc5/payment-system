package signals

import (
	"context"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/adapters/pubsub/kafka"
	"log/slog"
	"net/http"
	"sync"
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

func (s *Shutdown) Graceful(stopCh <-chan struct{}, httpServer *http.Server,
	consumer *kafka.Service,
	healthy *int32,
	wg *sync.WaitGroup) {
	ctx := context.Background()

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(ctx, s.serverShutdownTimeout)
	defer cancel()

	// all calls to /health will fail from now on
	atomic.StoreInt32(healthy, 0)

	s.logger.Info("shutting down", slog.Duration("timeout", s.serverShutdownTimeout))

	shutdownComplete := make(chan struct{})

	// stop OpenTelemetry tracer provider

	go func() {
		defer close(shutdownComplete)

		if httpServer != nil {
			s.logger.Info("http server is shutting down")
			if err := httpServer.Shutdown(ctx); err != nil {
				s.logger.Warn("HTTP server shutdown failed", slog.Any("error", err))
			}
		}

		if consumer != nil {
			s.logger.Info("consumer is shutting down")
			if err := consumer.Stop(); err != nil {
				s.logger.Warn("Consumer stop failed", slog.Any("error", err))
			}
		}
		wg.Wait()
	}()

	select {
	case <-shutdownComplete:
		s.logger.Info("shutdown complete")
	case <-ctx.Done():
		s.logger.Warn("shutdown timeout exceeded, forcing exit")
	}
}
