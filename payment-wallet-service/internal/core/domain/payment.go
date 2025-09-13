package domain

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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

type PaymentResultEvent struct {
	UserID        string         `json:"user_id"`
	TransactionID string         `json:"transaction_id"`
	Status        string         `json:"status"`
	Metadata      map[string]any `json:"metadata"`
}

func (cpr CreatePaymentRequest) Validate() error {
	err := validation.ValidateStruct(&cpr,
		validation.Field(&cpr.IdempotencyKey,
			validation.Required),
		validation.Field(&cpr.UserID,
			validation.Required),
		validation.Field(&cpr.Amount,
			validation.Required,
			validation.By(validAmount)),
		validation.Field(&cpr.ServiceID,
			validation.Required),
		validation.Field(&cpr.ClientNumber,
			validation.Required))
	if err != nil {
		return fmt.Errorf("error validating request %w", err)
	}

	return nil
}

func validAmount(value interface{}) error {
	v, _ := value.(int64)
	if v <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	return nil
}
