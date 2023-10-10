package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	Rest_latency prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		Rest_latency: *prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "REST_API",
			Name:      "rest_request_duration_seconds",
			Help:      "Duration of the rest api request",
			Buckets:   []float64{0.1, 0.15, 0.2, 0.25, 0.3},
		}, []string{"status", "method"}),
	}
	reg.MustRegister(m.Rest_latency)
	return m
}

func (s *APIServer) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &StatusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(recorder, r)
		duration := time.Since(start).Seconds()

		s.metrics.Rest_latency.WithLabelValues(strconv.Itoa(recorder.status), r.Method).Observe(duration)

	})
}

// to capture the status code
type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *StatusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
