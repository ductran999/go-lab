package fetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// slowServer sleeps ms then answers, unless the client went away.
func slowServer(ms int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := time.NewTimer(time.Duration(ms) * time.Millisecond)

		defer timer.Stop()

		select {
		case <-r.Context().Done():
			return
		case <-timer.C:
		}

		_, _ = w.Write([]byte("slow-done"))
	}))
}

func TestDoRespectsDeadline(t *testing.T) {
	t.Parallel()

	srv := slowServer(2000)

	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

	defer cancel()

	_, err := Do(ctx, srv.URL)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}

	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("err = %v, want it wrapped in ErrUpstream", err)
	}
}

func TestDoHedgedTakesTheFastFlight(t *testing.T) {
	t.Parallel()

	slow := slowServer(2000)

	defer slow.Close()

	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fast-win"))
	}))

	defer fast.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	start := time.Now()

	body, err := DoHedged(ctx, slow.URL, fast.URL, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("hedged: %v", err)
	}

	if body != "fast-win" {
		t.Fatalf("body = %q, want the hedge flight", body)
	}

	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Fatalf("took %v, hedge should cut the 2s tail", elapsed)
	}
}

func TestDoHedgedFastFailureSkipsHedge(t *testing.T) {
	t.Parallel()

	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer dead.Close()

	slow := slowServer(2000)

	defer slow.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	start := time.Now()

	_, err := DoHedged(ctx, dead.URL, slow.URL, time.Hour)
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream without waiting for the hedge", err)
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("took %v, fast failure must not wait out the hedge timer", elapsed)
	}
}

func TestDoFastestTakesFirstSuccess(t *testing.T) {
	t.Parallel()

	slow := slowServer(2000)

	defer slow.Close()

	mid := slowServer(500)

	defer mid.Close()

	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fast-win"))
	}))

	defer fast.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	start := time.Now()

	body, winner, err := DoFastest(ctx, slow.URL, mid.URL, fast.URL)
	if err != nil {
		t.Fatalf("race: %v", err)
	}

	if winner != 2 || body != "fast-win" {
		t.Fatalf("winner = %d %q, want flight 2 fast-win", winner, body)
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("took %v, the fast flight should win immediately", elapsed)
	}
}

func TestDoFastestJoinsAllFailures(t *testing.T) {
	t.Parallel()

	dead := func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
	}

	a := dead()

	defer a.Close()

	b := dead()

	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, _, err := DoFastest(ctx, a.URL, b.URL)
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("err = %v, want joined ErrUpstream", err)
	}
}
