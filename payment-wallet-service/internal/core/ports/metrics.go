package ports

import "time"

//go:generate mockgen -destination=../mocks/metrics_ports_mock.go -package=mocks -source=metrics.go

type Metrics interface {
	RecordPaymentCreated(paymentType, currency string, amount float64)
	RecordPaymentFailed(errorType, paymentType string)
	RecordPaymentProcessingTime(paymentType string, duration time.Duration)

	UpdateWalletBalance(walletID, currency string, balance float64)

	RecordTransactionStarted(transactionType string)
	RecordTransactionCompleted(transactionType string, success bool)

	RecordDatabaseOperationDuration(operation string, duration time.Duration)
	RecordExternalServiceCall(serviceName string, success bool, duration time.Duration)
}
