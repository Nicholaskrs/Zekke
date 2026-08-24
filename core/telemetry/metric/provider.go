package metric

import (
	"context"
	"net/http"
	"template-go/util/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Init sets up the MeterProvider with a Prometheus exporter.
// Returns an http.Handler to mount at /metrics, or nil if disabled.
func Init(ctx context.Context, config config.Config) (http.Handler, error) {
	if config.OpenTelemetryMetricsEnabled == false {
		return nil, nil
	}

	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("Zekke"),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(provider)

	if err := runtime.Start(runtime.WithMeterProvider(provider)); err != nil {
		return nil, err
	}

	return promhttp.Handler(), nil
}
