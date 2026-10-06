package breaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	errBoom  = errors.New("boom")
	errReset = errors.New("reset")
)

func TestOpensAfterThreshold(t *testing.T) {
	t.Parallel()

	b := New(3, time.Minute, nil)

	var runs atomic.Int64

	call := func() error {
		runs.Add(1)

		return errBoom
	}

	for range 3 {
		err := b.Call(call)
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v, want the real failure while closed", err)
		}
	}

	if b.State() != StateOpen {
		t.Fatalf("state = %v, want open after 3 consecutive", b.State())
	}

	err := b.Call(call)
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("err = %v, want ErrOpen without executing", err)
	}

	if got := runs.Load(); got != 3 {
		t.Fatalf("executions = %d, open circuit must not execute", got)
	}
}

func TestHalfOpenProbeClosesOnSuccess(t *testing.T) {
	t.Parallel()

	b := New(1, 30*time.Millisecond, nil)

	err := b.Call(func() error { return errBoom })
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	err = b.Call(func() error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("err = %v, want ErrOpen during cooldown", err)
	}

	time.Sleep(40 * time.Millisecond)

	err = b.Call(func() error { return nil })
	if err != nil {
		t.Fatalf("probe: %v", err)
	}

	if b.State() != StateClosed {
		t.Fatalf("state = %v, want closed after good probe", b.State())
	}
}

func TestFailedProbeReopens(t *testing.T) {
	t.Parallel()

	b := New(1, 20*time.Millisecond, nil)

	err := b.Call(func() error { return errBoom })
	if err == nil {
		t.Fatal("want the failure")
	}

	time.Sleep(30 * time.Millisecond)

	err = b.Call(func() error { return errBoom })
	if !errors.Is(err, errBoom) {
		t.Fatalf("probe err = %v, want the real failure", err)
	}

	if b.State() != StateOpen {
		t.Fatalf("state = %v, want open after failed probe", b.State())
	}

	err = b.Call(func() error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("err = %v, want ErrOpen in the new cooldown", err)
	}
}

func TestNonFailureNeverTrips(t *testing.T) {
	t.Parallel()

	b := New(2, time.Minute, func(error) bool { return false })

	for range 5 {
		err := b.Call(func() error { return errBoom })
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	}

	if b.State() != StateClosed {
		t.Fatalf("state = %v, uncounted errors must not trip", b.State())
	}
}

func TestRetryableStatusMatrix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		err  error
		want bool
	}{
		{errReset, true},
		{&StatusError{Status: 503, URL: "x"}, true},
		{&StatusError{Status: 429, URL: "x"}, true},
		{&StatusError{Status: 403, URL: "x"}, false},
		{&StatusError{Status: 400, URL: "x"}, false},
	}

	for _, c := range cases {
		got := RetryableStatus(c.err)
		if got != c.want {
			t.Errorf("RetryableStatus(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}

func TestHalfOpenAdmitsOneProbe(t *testing.T) {
	t.Parallel()

	b := New(1, 20*time.Millisecond, nil)

	err := b.Call(func() error { return errBoom })
	if err == nil {
		t.Fatal("want the failure")
	}

	time.Sleep(30 * time.Millisecond)

	release := make(chan struct{})

	var runs atomic.Int64

	var wg sync.WaitGroup

	errs := make([]error, 5)

	for i := range 5 {
		wg.Go(func() {
			errs[i] = b.Call(func() error {
				runs.Add(1)
				<-release

				return nil
			})
		})
	}

	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	if got := runs.Load(); got != 1 {
		t.Fatalf("probe executions = %d, want exactly 1", got)
	}

	rejected := 0

	for _, err := range errs {
		if errors.Is(err, ErrOpen) {
			rejected++
		}
	}

	if rejected != 4 {
		t.Fatalf("rejected = %d, want the 4 non-probes refused", rejected)
	}
}
