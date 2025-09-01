package http

import (
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/gorilla/mux"
	"log/slog"
)

type deps struct {
	balanceSvc *mocks.MockBalanceService
	paymentSvc *mocks.MockPaymentService
}

func NewMockServer(deps *deps) *Server {
	return &Server{
		port:           5555,
		logger:         slog.Default(),
		router:         mux.NewRouter(),
		paymentService: deps.paymentSvc,
		balanceService: deps.balanceSvc,
	}
}
