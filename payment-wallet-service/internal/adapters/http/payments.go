package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
)

func (s *Server) createPaymentHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		s.ErrorResponse(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("cannot unmarshal request", slog.Any("error", err))
		s.ErrorResponse(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	req.UserID = userID

	if err := req.Validate(); err != nil {
		s.logger.Error("validation error", slog.Any("error", err))
		s.ErrorResponse(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	err := s.paymentService.Create(context.TODO(), req)
	if err != nil {
		s.logger.Error("cannot create payment", slog.Any("error", err))

		s.ErrorResponse(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
