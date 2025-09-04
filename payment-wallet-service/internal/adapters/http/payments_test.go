package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/domain"
	"github.com/emiliocc5/payment-system/payment-wallet-service/internal/core/ports/mocks"
	"github.com/golang/mock/gomock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_createPaymentHandler(t *testing.T) {
	tests := []struct {
		name                 string
		userID               string
		requestBody          interface{}
		paymentServiceError  error
		validateError        error
		expectedStatusCode   int
		expectedErrorMessage string
		paymentServiceTimes  int
	}{
		{
			name:   "Success - Valid request",
			userID: "user123",
			requestBody: domain.CreatePaymentRequest{
				UserID:         "user123",
				ClientNumber:   "client-number",
				ServiceID:      "service-id",
				Amount:         10,
				IdempotencyKey: "test-idempotency-key",
			},
			expectedStatusCode:  http.StatusCreated,
			paymentServiceTimes: 1,
		},
		{
			name:                 "Error - Missing User ID",
			userID:               "",
			requestBody:          domain.CreatePaymentRequest{Amount: 10050},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "unauthorized",
			paymentServiceTimes:  0,
		},
		{
			name:                "Error - Invalid JSON body",
			userID:              "user123",
			requestBody:         "invalid-json",
			expectedStatusCode:  http.StatusBadRequest,
			paymentServiceTimes: 0,
		},
		{
			name:   "Error - Payment service fails",
			userID: "user123",
			requestBody: domain.CreatePaymentRequest{
				UserID:         "user123",
				ClientNumber:   "client-number",
				ServiceID:      "service-id",
				Amount:         10,
				IdempotencyKey: "test-idempotency-key",
			},
			paymentServiceError:  errors.New("service error"),
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "service error",
			paymentServiceTimes:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockPaymentSvc := mocks.NewMockPaymentService(ctrl)
			mockPaymentSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				Return(tt.paymentServiceError).Times(tt.paymentServiceTimes)

			server := &Server{
				logger:         slog.Default(),
				paymentService: mockPaymentSvc,
			}

			var bodyReader *bytes.Reader
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					bodyReader = bytes.NewReader([]byte(str))
				} else {
					jsonBody, _ := json.Marshal(tt.requestBody)
					bodyReader = bytes.NewReader(jsonBody)
				}
			} else {
				bodyReader = bytes.NewReader([]byte("{}"))
			}

			req := httptest.NewRequest(http.MethodPost, "/payments", bodyReader)
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			w := httptest.NewRecorder()

			server.createPaymentHandler(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			if tt.expectedErrorMessage != "" {
				body := w.Body.String()
				if !strings.Contains(body, tt.expectedErrorMessage) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrorMessage, body)
				}
			}
		})
	}
}
