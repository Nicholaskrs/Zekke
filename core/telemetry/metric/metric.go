package metric

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func Meter(name string) metric.Meter {
	return otel.Meter(name)
}
