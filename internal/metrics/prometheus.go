package metrics

import (
	"audit-proxy-gateway/pkg/logger"
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Prometheus metrics
var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	HttpResponseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_status",
			Help: "HTTP response status codes",
		},
		[]string{"path", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpResponseStatus)
}

func TrackMetrics(c *fiber.Ctx) error {
	path := c.Path()

	if path == "/metrics" {
		return c.Next()
	}

	start := time.Now()
	method := c.Method()
	err := c.Next()

	responseStatusCode := strconv.Itoa(c.Response().StatusCode())

	duration := time.Since(start).Seconds()
	HttpRequestsTotal.WithLabelValues(path, method).Inc()
	HttpRequestDuration.WithLabelValues(path, method).Observe(duration)
	HttpResponseStatus.WithLabelValues(path, method, responseStatusCode).Inc()

	logger.GetLogger().Infof("Request to %s %s took %f seconds", method, path, duration)

	return err
}

// AdaptedResponseWriter wraps a fiber.Response to implement http.ResponseWriter
type AdaptedResponseWriter struct {
	header http.Header
	body   *bytes.Buffer
	status int
}

func (arw *AdaptedResponseWriter) Header() http.Header {
	return arw.header
}

func (arw *AdaptedResponseWriter) Write(b []byte) (int, error) {
	return arw.body.Write(b)
}

func (arw *AdaptedResponseWriter) WriteHeader(statusCode int) {
	arw.status = statusCode
}

// AdaptFiberRequestToHTTP converts a *fiber.Ctx to *http.Request
func AdaptFiberRequestToHTTP(c *fiber.Ctx) *http.Request {
	req := new(http.Request)
	req.Method = c.Method()
	req.URL, _ = url.ParseRequestURI(c.OriginalURL())
	req.Body = io.NopCloser(bytes.NewReader(c.Body()))
	req.Header = make(http.Header)
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
	return req
}

// Prometheus metrics endpoint
func MetricsEndpoint(c *fiber.Ctx) error {
	w := &AdaptedResponseWriter{
		header: make(http.Header),
		body:   new(bytes.Buffer),
	}

	r := AdaptFiberRequestToHTTP(c)
	promhttp.Handler().ServeHTTP(w, r)

	for k, v := range w.Header() {
		for _, val := range v {
			c.Set(k, val)
		}
	}

	c.Status(w.status)
	return c.Send(w.body.Bytes())
}
