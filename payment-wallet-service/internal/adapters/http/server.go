package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports"
	"github.com/gorilla/mux"

	"net/http"
	"sync/atomic"
	"time"
)

type ServerConfig struct {
	Port           int
	PaymentService ports.PaymentService
	BalanceService ports.BalanceService
}

type Server struct {
	port           int
	logger         *slog.Logger
	router         *mux.Router
	handler        http.Handler
	paymentService ports.PaymentService
	balanceService ports.BalanceService
}

var (
	_healthy int32
)

func NewServer(cfg *ServerConfig, logger *slog.Logger) *Server {
	return &Server{
		port:           cfg.Port,
		logger:         logger,
		router:         mux.NewRouter(),
		paymentService: cfg.PaymentService,
		balanceService: cfg.BalanceService,
	}
}

func (s *Server) registerHandlers() {
	sub := s.router.PathPrefix("/v1").Subrouter()
	sub.HandleFunc("/health", s.healthHandler).Methods(http.MethodGet)
	sub.HandleFunc("/payments", s.createPaymentHandler).Methods(http.MethodPost)
}

func (s *Server) start() *http.Server {
	if s.port == 0 {
		return nil
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		s.logger.Info("starting server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil {
			s.logger.Error("HTTP server crashed", slog.Any("error", err))
		}
	}()

	return srv
}

func (s *Server) ListenAndServe(ctx context.Context) (*http.Server, *int32) {
	s.registerHandlers()
	s.handler = s.router

	srv := s.start()

	atomic.StoreInt32(&_healthy, 1)

	//go s.ps.Listen(ctx)

	return srv, &_healthy
}
