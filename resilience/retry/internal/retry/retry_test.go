package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Sentinel flakes: err113 forbids dynamic errors even in tests.
var (
	errBlip       = errors.New("blip")
	errDown       = errors.New("down")
	errAlwaysDown = errors.New("always down")
)

func TestCallSucceedsAfterFlakes(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Call(context.Background(), 5, 5*time.Millisecond, false, nil, func(context.Context) error {
		calls++

		if calls < 3 {
			return errBlip
		}

		return nil
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}

	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (2 flakes then success)", calls)
	}
}

func TestCallExhaustsAttempts(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Call(context.Background(), 3, time.Millisecond, false, nil, func(context.Context) error {
		calls++

		return errAlwaysDown
	})
	if !errors.Is(err, ErrExhausted) {
		t.Fatalf("err = %v, want ErrExhausted", err)
	}

	if calls != 3 {
		t.Fatalf("calls = %d, want exactly 3 attempts", calls)
	}
}

func TestCallAbortsOnCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()

	err := Call(ctx, 10, time.Hour, false, nil, func(context.Context) error {
		return errDown
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled unwrapped", err)
	}

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("took %v, cancel must abort the hour-long sleep", elapsed)
	}
}

func TestCallJitterStillSucceeds(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Call(context.Background(), 5, time.Millisecond, true, nil, func(context.Context) error {
		calls++

		if calls < 3 {
			return errBlip
		}

		return nil
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}

	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (2 flakes then success)", calls)
	}
}

func TestCallStopsOnNonRetryable(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Call(context.Background(), 5, time.Millisecond, true, RetryableStatus, func(context.Context) error {
		calls++

		return &StatusError{Status: 403, URL: "/deny"}
	})

	var status *StatusError
	if !errors.As(err, &status) || status.Status != 403 {
		t.Fatalf("err = %v, want the original 403 unretried", err)
	}

	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (dead errors stop at once)", calls)
	}
}

func TestCallRetriesRetryableStatus(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Call(context.Background(), 5, time.Millisecond, true, RetryableStatus, func(context.Context) error {
		calls++

		if calls < 3 {
			return &StatusError{Status: 503, URL: "/flaky"}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}

	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (2x503 then success)", calls)
	}
}

func TestBackoffBounds(t *testing.T) {
	t.Parallel()

	if got := backoff(100*time.Millisecond, 2, false); got != 400*time.Millisecond {
		t.Fatalf("plain backoff = %v, want 400ms", got)
	}

	if got := backoff(time.Hour, 100, false); got != maxDelay {
		t.Fatalf("uncapped backoff = %v, want %v", got, maxDelay)
	}

	for range 50 {
		got := backoff(100*time.Millisecond, 0, true)
		if got < 0 || got > 100*time.Millisecond {
			t.Fatalf("jittered backoff = %v, want within [0, 100ms]", got)
		}
	}
}
