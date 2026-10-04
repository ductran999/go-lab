package golab

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetricsQueryFiltersAndCaps(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("http_requests_total{route=\"/a\"} 1\nhttp_requests_total{route=\"/b\"} 2\ngo_goroutines 5\n"))
	}))

	defer srv.Close()

	out, err := MetricsQuery(srv.URL, "http_requests_total")
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if strings.Count(out, "\n") != 1 || !strings.Contains(out, "/a") {
		t.Fatalf("want exactly the 2 matching lines, got:\n%s", out)
	}

	_, err = MetricsQuery(srv.URL, "")
	if err == nil {
		t.Fatal("empty pattern must fail (would dump everything)")
	}

	out, err = MetricsQuery(srv.URL, "nothing-here")
	if err != nil || !strings.Contains(out, "no lines match") {
		t.Fatalf("no-match = %q, %v", out, err)
	}
}

func TestRetryStatsDumpsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stats" {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_, _ = w.Write([]byte(`{"hits":20}`))
	}))

	defer srv.Close()

	out, err := RetryStats(srv.URL)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}

	if !strings.Contains(out, `"hits":20`) {
		t.Fatalf("body = %q", out)
	}
}

func TestBenchRunBoundsAndStub(t *testing.T) {
	t.Parallel()

	_, err := BenchRun(t.TempDir(), 0, 10)
	if err == nil {
		t.Fatal("n=0 must fail")
	}

	dir := t.TempDir()
	stub := "#!/bin/sh\necho \"stub bench $1 $2\"\n"

	//nolint:gosec // stub must execute for the test
	err = os.WriteFile(filepath.Join(dir, "bench.sh"), []byte(stub), 0o755)
	if err != nil {
		t.Fatalf("stub: %v", err)
	}

	out, err := BenchRun(dir, 50, 10)
	if err != nil {
		t.Fatalf("stub bench: %v", err)
	}

	if out != "stub bench 50 10" {
		t.Fatalf("out = %q", out)
	}
}
