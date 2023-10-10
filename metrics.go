package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
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

// middleware for rest api latency
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

//----------------------------------------------------------------------
//gRPC metrics
//----------------------------------------------------------------------

type grpcMetrics struct {
	createAccount_latency prometheus.HistogramVec
}

func NewGRPCMetrics(reg prometheus.Registerer) *grpcMetrics {
	m := &grpcMetrics{
		*prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "create_account_latency_seconds",
				Help: "duration of the CreateAccount gRPC",
			},
			[]string{"method", "status"},
		),
	}
	reg.MustRegister(m.createAccount_latency)
	return m
}

// UnaryInterceptor act as middleware for latency capture
func (m *grpcMetrics) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
	}
	m.createAccount_latency.WithLabelValues(info.FullMethod, status).Observe(duration.Seconds())
	return resp, err
}

func StartGRPCMetricsServer() *grpcMetrics {

	reg := prometheus.NewRegistry()
	metrics := NewGRPCMetrics(reg)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":9092", nil)
	return metrics

}
