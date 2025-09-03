package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusMetrics struct {
	paymentsCreated           *prometheus.CounterVec
	paymentsFailed            *prometheus.CounterVec
	paymentAmount             *prometheus.HistogramVec
	paymentProcessingDuration *prometheus.HistogramVec
	walletBalance             *prometheus.GaugeVec
	transactionsStarted       *prometheus.CounterVec
	transactionsCompleted     *prometheus.CounterVec
	dbOperationDuration       *prometheus.HistogramVec
	externalServiceCalls      *prometheus.CounterVec
	externalServiceDuration   *prometheus.HistogramVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		paymentsCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_payments_created_total",
				Help: "Total number of payments created successfully",
			},
			[]string{"payment_type", "currency"},
		),

		paymentsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "payment_wallet_payments_failed_total",
				Help: "Total number of failed payments",
			},
			[]string{"error_type", "payment_type"},
		),

		paymentAmount: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_wallet_payment_amount",
				Help:    "Payment amounts processed",
				Buckets: []float64{1, 10, 50, 100, 500, 1000, 5000, 10000},
			},
			[]string{"currency", "payment_type"},
		),

		paymentProcessingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "payment_wallet_payment_processing_duration_seconds",
				Help:    "Time taken to process payments",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"payment_type"},
		),

		walletBalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "payment_wallet_balance_total",
				Help: "Current wallet balances",
			},
			[]string{"wallet_id", "currency"},
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

	// Registrar todas las métricas con Prometheus
	prometheus.MustRegister(
		metrics.paymentsCreated,
		metrics.paymentsFailed,
		metrics.paymentAmount,
		metrics.paymentProcessingDuration,
		metrics.walletBalance,
		metrics.transactionsStarted,
		metrics.transactionsCompleted,
		metrics.dbOperationDuration,
		metrics.externalServiceCalls,
		metrics.externalServiceDuration,
	)

	return metrics
}

func (m *PrometheusMetrics) RecordPaymentCreated(paymentType, currency string, amount float64) {
	m.paymentsCreated.WithLabelValues(paymentType, currency).Inc()
	m.paymentAmount.WithLabelValues(currency, paymentType).Observe(amount)
}

func (m *PrometheusMetrics) RecordPaymentFailed(errorType, paymentType string) {
	m.paymentsFailed.WithLabelValues(errorType, paymentType).Inc()
}

func (m *PrometheusMetrics) RecordPaymentProcessingTime(paymentType string, duration time.Duration) {
	m.paymentProcessingDuration.WithLabelValues(paymentType).Observe(duration.Seconds())
}

func (m *PrometheusMetrics) UpdateWalletBalance(walletID, currency string, balance float64) {
	m.walletBalance.WithLabelValues(walletID, currency).Set(balance)
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
