// Package retry reruns fallible work with exponential backoff and
// optional full jitter: sleep = rand[0, min(cap, base*2^n)]. The
// jitter is the point: N clients failing on the same blip retry in
// lockstep without it (a second thundering herd aimed at a server
// that just recovered), and spread out with it.
//
// Stdlib-only on purpose (zero-dep lab). Production: cenkaler/backoff
// maps 1:1 — InitialInterval=base, Multiplier=2, MaxInterval=cap,
// RandomizationFactor=±jitter, WithMaxRetries=attempts,
// WithContext=ctx, Permanent=non-retryable.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

// ErrExhausted wraps the last failure once attempts run out.
var ErrExhausted = errors.New("retries exhausted")

// maxDelay caps a single sleep: backoff must never outgrow the
// caller's patience (or its deadline).
const maxDelay = 5 * time.Second

// Call runs fn until it succeeds or attempts run out. The first try
// is immediate; sleeps grow base, 2*base, 4*base... A cancelled
// context aborts the sleep and returns ctx.Err unwrapped, so
// errors.Is(err, context.Canceled) keeps working. Retryable or not
// is fn's contract: return nil on success, an error worth retrying
// otherwise (never retry a 400, always bound the attempts).
func Call(ctx context.Context, attempts int, base time.Duration, jitter bool, fn func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}

	var err error

	for n := range attempts {
		err = fn(ctx)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if n == attempts-1 {
			break
		}

		timer := time.NewTimer(backoff(base, n, jitter))

		select {
		case <-ctx.Done():
			timer.Stop()

			return ctx.Err()
		case <-timer.C:
		}
	}

	return fmt.Errorf("%w after %d attempts: %w", ErrExhausted, attempts, err)
}

// backoff returns base*2^n capped at maxDelay, or a uniform draw
// from [0, capped] under jitter (AWS "full jitter": the average
// sleep halves, the spikes disappear).
func backoff(base time.Duration, n int, jitter bool) time.Duration {
	capped := base << n
	if capped <= 0 || capped > maxDelay {
		capped = maxDelay
	}

	if !jitter {
		return capped
	}

	return time.Duration(rand.Int64N(int64(capped) + 1))
}
