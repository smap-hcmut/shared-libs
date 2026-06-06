// Package metrics provides reusable Prometheus instrumentation for SMAP Go
// services. The goal is to expose a uniform HTTP RED metric set with the
// service name attached, so the SMAP overview Grafana dashboard can split
// every panel by service without per-service configuration drift.
package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// registerOnce guards against duplicate registration when a service runs more
// than one Gin engine in the same process (e.g. internal + public).
var registerOnce sync.Once

var (
	httpRequestsTotal *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
	httpInFlight      *prometheus.GaugeVec
)

func ensureRegistered() {
	registerOnce.Do(func() {
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "smap_http_requests_total",
				Help: "Total HTTP requests handled by an SMAP Go service.",
			},
			[]string{"service", "method", "route", "status"},
		)
		httpDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "smap_http_request_duration_seconds",
				Help:    "HTTP request latency in seconds, observed at the application.",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
			},
			[]string{"service", "method", "route"},
		)
		httpInFlight = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "smap_http_in_flight_requests",
				Help: "Currently in-flight HTTP requests, per service.",
			},
			[]string{"service"},
		)

		prometheus.MustRegister(httpRequestsTotal, httpDuration, httpInFlight)
	})
}

// GinMiddleware returns a Gin middleware that records request count, latency
// and in-flight gauge for every request. The `service` label is fixed at
// middleware creation time so labels match the deployment identity, not any
// header an upstream proxy may inject.
//
// Route templates are taken from gin.Context.FullPath() so cardinality stays
// bounded — `/foo/:id` keeps one label set regardless of how many ids hit it.
// Requests that don't match a route (404) report route="<unmatched>".
func GinMiddleware(service string) gin.HandlerFunc {
	ensureRegistered()
	return func(c *gin.Context) {
		// Don't record the scrape endpoint itself.
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		start := time.Now()
		httpInFlight.WithLabelValues(service).Inc()
		c.Next()
		httpInFlight.WithLabelValues(service).Dec()

		route := c.FullPath()
		if route == "" {
			route = "<unmatched>"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(service, c.Request.Method, route, status).Inc()
		httpDuration.WithLabelValues(service, c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

// MountMetrics registers GET /metrics on the given Gin engine, serving the
// default Prometheus registry via promhttp.
func MountMetrics(r *gin.Engine) {
	ensureRegistered()
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})))
}

// PromHTTPHandler returns a bare http.Handler for services that don't use Gin
// for their metrics endpoint (e.g. when serving /metrics from a separate
// admin port).
func PromHTTPHandler() http.Handler {
	ensureRegistered()
	return promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
}
