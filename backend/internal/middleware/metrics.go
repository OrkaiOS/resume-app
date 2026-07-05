package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds the Prometheus collectors used to instrument HTTP traffic.
// Construct one with NewMetrics and pass its Handler() to the Gin middleware
// chain. Collectors are registered on construction (no package-level state).
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestErrors   *prometheus.CounterVec
}

// NewMetrics constructs a Metrics, registers its collectors with the default
// Prometheus registry, and returns it.
func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests by method, route, and status.",
			},
			[]string{"method", "route", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds by method and route.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route"},
		),
		requestErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_errors_total",
				Help: "Total number of HTTP requests that returned a client or server error (status >= 400).",
			},
			[]string{"method", "route", "status"},
		),
	}
	prometheus.MustRegister(m.requestsTotal, m.requestDuration, m.requestErrors)
	return m
}

// Handler returns a Gin middleware that records request count, latency, and
// error rate for each request.
func (m *Metrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		elapsed := time.Since(start).Seconds()

		m.requestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		m.requestDuration.WithLabelValues(c.Request.Method, route).Observe(elapsed)
		if c.Writer.Status() >= 400 {
			m.requestErrors.WithLabelValues(c.Request.Method, route, status).Inc()
		}
	}
}
