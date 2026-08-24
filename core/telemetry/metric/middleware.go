package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
)

// mustInitInstruments creates the counter/histogram once, reused across requests.
func mustInitInstruments() {
	m := Meter("zekke/http")

	var err error
	requestCounter, err = m.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total HTTP requests"),
	)
	if err != nil {
		panic(err)
	}

	requestDuration, err = m.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		panic(err)
	}
}

func GinMiddleware() gin.HandlerFunc {
	mustInitInstruments()

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		attrs := metric.WithAttributes(
			attribute.String("method", c.Request.Method),
			attribute.String("route", c.FullPath()),
			attribute.String("status", strconv.Itoa(c.Writer.Status())),
		)

		requestCounter.Add(c.Request.Context(), 1, attrs)
		requestDuration.Record(c.Request.Context(), duration, attrs)
	}
}
