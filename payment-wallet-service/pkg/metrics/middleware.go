package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_wallet_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_wallet_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	activeRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "payment_wallet_active_requests",
			Help: "Number of active HTTP requests",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeRequests)
}

func Middleware(next http.Handler, log slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/metrics" || r.URL.Path == "/v1/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		activeRequests.Inc()
		defer activeRequests.Dec()

		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()
		endpoint := r.URL.Path
		method := r.Method
		statusCode := strconv.Itoa(ww.statusCode)

		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()

		log.Debug("HTTP request completed",
			slog.String("method", method),
			slog.String("endpoint", endpoint),
			slog.String("status", statusCode),
			slog.Float64("duration_seconds", duration),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
