package ports

import "time"

//go:generate mockgen -destination=./mocks/metrics_ports_mock.go -package=mocks -source=metrics.go

type Metrics interface {
	UpdateWalletBalance(walletID, currency string, balance float64)

	RecordTransactionIdempotent(transactionType string)
	RecordTransactionStarted(transactionType string)
	RecordTransactionCompleted(transactionType string, success bool)
	RecordTransactionAmount(transactionType string, amount float64)
	RecordTransactionProcessingTime(transactionType string, duration time.Duration)

	RecordDatabaseOperationDuration(operation string, duration time.Duration)
	RecordExternalServiceCall(serviceName string, success bool, duration time.Duration)
}
