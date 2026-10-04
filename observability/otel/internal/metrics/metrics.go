// Package metrics exposes Prometheus counters + histograms for the
// demo services: requests total and duration, labeled by route.
// One /metrics scrape endpoint per binary (no collector hop).
package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// Recorder counts requests and durations. Built once by Setup.
type Recorder struct {
	requests metric.Int64Counter
	duration metric.Float64Histogram
}

// Current is the process recorder (set by Setup).
var Current *Recorder

// Setup registers the Prometheus exporter as global meter provider and
// returns the scrape handler. Call once in main.
func Setup() (http.Handler, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("metrics: exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	meter := provider.Meter("demo")

	requests, err := meter.Int64Counter("http_requests_total",
		metric.WithDescription("total requests by route and status"))
	if err != nil {
		return nil, fmt.Errorf("metrics: counter: %w", err)
	}

	duration, err := meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("request duration by route"))
	if err != nil {
		return nil, fmt.Errorf("metrics: histogram: %w", err)
	}

	Current = &Recorder{requests: requests, duration: duration}

	return promhttp.Handler(), nil
}

// Observe records one request (route pattern, status, seconds).
func (r *Recorder) Observe(ctx context.Context, route string, status int, seconds float64) {
	if r == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("route", route),
		attribute.Int("status", status),
	}

	r.requests.Add(ctx, 1, metric.WithAttributes(attrs...))
	r.duration.Record(ctx, seconds, metric.WithAttributes(attrs...))
}
