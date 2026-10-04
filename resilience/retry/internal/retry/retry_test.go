package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCallSucceedsAfterFlakes(t *testing.T) {
	t.Parallel()

	calls := 0

	err := Call(context.Background(), 5, 5*time.Millisecond, false, func(context.Context) error {
		calls++

		if calls < 3 {
			return errors.New("blip")
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

	err := Call(context.Background(), 3, time.Millisecond, false, func(context.Context) error {
		calls++

		return errors.New("always down")
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

	err := Call(ctx, 10, time.Hour, false, func(context.Context) error {
		return errors.New("down")
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

	err := Call(context.Background(), 5, time.Millisecond, true, func(context.Context) error {
		calls++

		if calls < 3 {
			return errors.New("blip")
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
