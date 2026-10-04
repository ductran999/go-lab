// Package metrics exposes Prometheus counters + histograms with
// exemplars: every sample carries its trace.id, so a graph spike
// jumps straight to the Jaeger waterfall. Native client_golang
// (OTel SDK v1.44 doesn't attach exemplars itself).
package metrics

import (
	"context"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
)

var (
	requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "total requests by route and status",
		},
		[]string{"route", "status"},
	)
	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "request duration by route",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "status"},
	)
	inflight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_inflight_requests",
			Help: "requests currently being served by route",
		},
		[]string{"route"},
	)
	respSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_response_size_bytes",
			Help:       "response size by route",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"route"},
	)
)

func init() {
	prometheus.MustRegister(requests, duration, inflight, respSize)
}

// Handler serves /metrics (own ports :2112/:2113, no auth).
// EnableOpenMetrics: exemplars are only emitted in OpenMetrics format.
func Handler() http.Handler {
	reg := prometheus.DefaultGatherer
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

// Recorder is kept for API shape (no state: instruments are global).
type Recorder struct{}

// Current is the process recorder (set by Setup for compatibility).
var Current = &Recorder{}

// Setup exists so mains keep one call; instruments register at init.
func Setup() (http.Handler, error) {
	Current = &Recorder{}

	return Handler(), nil
}

// Track marks one request in-flight: call at entry, defer the
// returned done until the response is written. A gauge that never
// returns to 0 after load is a leak (forgotten Dec, stuck handler).
func (r *Recorder) Track(route string) func() {
	noop := func() {}

	if r == nil {
		return noop
	}

	g := inflight.With(prometheus.Labels{"route": route})
	g.Inc()

	return g.Dec
}

// Observe records one finished request: counter + histogram carry the
// trace exemplar, summary tracks response size (local quantiles only —
// summaries never aggregate across replicas).
func (r *Recorder) Observe(ctx context.Context, route string, status int, seconds float64, respBytes int) {
	if r == nil {
		return
	}

	sc := trace.SpanFromContext(ctx).SpanContext()
	traceID := sc.TraceID().String()
	labels := prometheus.Labels{"route": route, "status": strconv.Itoa(status)}
	exemplar := prometheus.Labels{"trace_id": traceID}

	if adder, ok := requests.With(labels).(prometheus.ExemplarAdder); ok {
		adder.AddWithExemplar(1, exemplar)
	}

	if observer, ok := duration.With(labels).(prometheus.ExemplarObserver); ok {
		observer.ObserveWithExemplar(seconds, exemplar)
	}

	respSize.With(prometheus.Labels{"route": route}).Observe(float64(respBytes))
}
