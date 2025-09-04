package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusMetrics struct {
	walletBalance                 *prometheus.GaugeVec
	transactionsIdempotent        *prometheus.CounterVec
	transactionsStarted           *prometheus.CounterVec
	transactionsCompleted         *prometheus.CounterVec
	transactionAmount             *prometheus.CounterVec
	transactionProcessingDuration *prometheus.HistogramVec
	dbOperationDuration           *prometheus.HistogramVec
	externalServiceCalls          *prometheus.CounterVec
	externalServiceDuration       *prometheus.HistogramVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		walletBalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "payment_wallet_balance_total",
				Help: "Current wallet balances",
			},
			[]string{"wallet_id", "currency"},
		),
		transactionsIdempotent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_transactions_idempotent_total",
				Help: "Total number of transactions idempotent",
			},
			[]string{"transaction_type"},
		),
		transactionsStarted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_transactions_started_total",
				Help: "Total number of transactions started",
			},
			[]string{"transaction_type"},
		),
		transactionsCompleted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_transactions_completed_total",
				Help: "Total number of transactions completed",
			},
			[]string{"transaction_type", "success"},
		),
		transactionAmount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_transactions_amount",
				Help: "Transaction amounts processed",
			},
			[]string{"transaction_type", "currency"},
		),
		transactionProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_wallet_transactions_processing_duration_seconds",
				Help:    "Time taken to process transactions",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"transaction_type"},
		),

		dbOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_wallet_db_operation_duration_seconds",
				Help:    "Duration of database operations",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"operation"},
		),

		externalServiceCalls: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_external_service_calls_total",
				Help: "Total number of external service calls",
			},
			[]string{"service_name", "success"},
		),

		externalServiceDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_wallet_external_service_duration_seconds",
				Help:    "Duration of external service calls",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"service_name"},
		),
	}

	prometheus.MustRegister(
		metrics.walletBalance,
		metrics.transactionsStarted,
		metrics.transactionsCompleted,
		metrics.transactionAmount,
		metrics.transactionsIdempotent,
		metrics.transactionProcessingDuration,
		metrics.dbOperationDuration,
		metrics.externalServiceCalls,
		metrics.externalServiceDuration,
	)

	return metrics
}

func (m *PrometheusMetrics) UpdateWalletBalance(walletID, currency string, balance float64) {
	m.walletBalance.WithLabelValues(walletID, currency).Set(balance)
}

func (m *PrometheusMetrics) RecordTransactionIdempotent(transactionType string) {
	m.transactionsIdempotent.WithLabelValues(transactionType).Inc()
}

func (m *PrometheusMetrics) RecordTransactionStarted(transactionType string) {
	m.transactionsStarted.WithLabelValues(transactionType).Inc()
}

func (m *PrometheusMetrics) RecordTransactionCompleted(transactionType string, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	m.transactionsCompleted.WithLabelValues(transactionType, successStr).Inc()
}

func (m *PrometheusMetrics) RecordTransactionAmount(transactionType string, amount float64) {
	m.transactionAmount.WithLabelValues(transactionType).Add(amount)
}

func (m *PrometheusMetrics) RecordTransactionProcessingTime(transactionType string, duration time.Duration) {
	m.transactionProcessingDuration.WithLabelValues(transactionType).Observe(duration.Seconds())
}

func (m *PrometheusMetrics) RecordDatabaseOperationDuration(operation string, duration time.Duration) {
	m.dbOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

func (m *PrometheusMetrics) RecordExternalServiceCall(serviceName string, success bool, duration time.Duration) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	m.externalServiceCalls.WithLabelValues(serviceName, successStr).Inc()
	m.externalServiceDuration.WithLabelValues(serviceName).Observe(duration.Seconds())
}
