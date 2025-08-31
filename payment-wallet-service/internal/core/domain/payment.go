package domain

import (
	"time"
)

type CreatePaymentRequest struct {
	UserID         string `json:"user_id"`
	ClientNumber   string `json:"client_number"`
	ServiceID      string `json:"service_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

type Payment struct {
	ID             string    `json:"id"`
	IdempotencyKey string    `json:"idempotency_key"`
	UserID         string    `json:"user_id"`
	Amount         int64     `json:"amount"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ServiceID      string    `json:"service_id"`
	ClientNumber   string    `json:"client_number"`
}

type PaymentInitiatedEvent struct {
	UserID        string `json:"user_id"`
	ClientNumber  string `json:"client_number"`
	ServiceID     string `json:"service_id"`
	Amount        int64  `json:"amount"`
	TransactionID string `json:"transaction_id"`
}

type PaymentResultEvent struct{}

// TODO add validations
func (cpr *CreatePaymentRequest) Validate() error {
	return nil
}
