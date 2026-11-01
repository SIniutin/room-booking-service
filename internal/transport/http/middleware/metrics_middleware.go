package middleware

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/SIniutin/room-booking-service/internal/metrics"
)

func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &statusRecorder{ResponseWriter: w, body: &bytes.Buffer{}, status: 200}
			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rec.status)

			m.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
			m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		})
	}
}
