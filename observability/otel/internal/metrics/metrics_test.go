package metrics

import (
	"context"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel/trace"
)

// A KNOWN trace id, no SDK needed.
func ctxWithTrace() context.Context {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:  [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	})

	return trace.ContextWithSpanContext(context.Background(), sc)
}

func TestExemplarSurfaces(t *testing.T) {
	t.Parallel()

	if _, err := Setup(); err != nil {
		t.Fatalf("setup: %v", err)
	}

	Current.Observe(ctxWithTrace(), "/test", 200, 0.05, 512)

	plain, err := testutil.CollectAndFormat(requests, expfmt.TypeTextPlain, "http_requests_total")
	if err != nil {
		t.Fatalf("collect plain: %v", err)
	}

	t.Logf("plain counter:\n%s", plain)

	if !strings.Contains(string(plain), `http_requests_total{route="/test",status="200"} 1`) {
		t.Errorf("counter sample missing entirely (assertion failed?):\n%s", plain)
	}

	cout, err := testutil.CollectAndFormat(requests, expfmt.TypeOpenMetrics, "http_requests_total")
	if err != nil {
		t.Fatalf("collect counter: %v", err)
	}

	t.Logf("counter exposition:\n%s", cout)

	if !strings.Contains(string(cout), "0102030405060708090a0b0c0d0e0f10") {
		t.Errorf("counter exposition lacks trace exemplar:\n%s", cout)
	}

	dout, err := testutil.CollectAndFormat(duration, expfmt.TypeOpenMetrics, "http_request_duration_seconds")
	if err != nil {
		t.Fatalf("collect histogram: %v", err)
	}

	t.Logf("histogram exposition:\n%s", dout)

	if !strings.Contains(string(dout), "0102030405060708090a0b0c0d0e0f10") {
		t.Errorf("histogram exposition lacks trace exemplar:\n%s", dout)
	}

	sout, err := testutil.CollectAndFormat(respSize, expfmt.TypeTextPlain, "http_response_size_bytes")
	if err != nil {
		t.Fatalf("collect summary: %v", err)
	}

	t.Logf("summary exposition:\n%s", sout)

	if !strings.Contains(string(sout), `http_response_size_bytes{route="/test",quantile="0.5"}`) {
		t.Errorf("summary exposition lacks quantile series:\n%s", sout)
	}
}

func TestInflightReturnsToZero(t *testing.T) {
	t.Parallel()

	if _, err := Setup(); err != nil {
		t.Fatalf("setup: %v", err)
	}

	g := inflight.With(map[string]string{"route": "/gauge-test"})

	done := Current.Track("/gauge-test")

	if got := testutil.ToFloat64(g); got != 1 {
		t.Fatalf("inflight during request = %v, want 1", got)
	}

	done()

	if got := testutil.ToFloat64(g); got != 0 {
		t.Errorf("inflight after request = %v, want 0 (leak?)", got)
	}
}
