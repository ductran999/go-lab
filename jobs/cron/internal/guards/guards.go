// Package guards wraps job funcs with cross-cutting concerns:
// overlap locks (SkipIfStillRunning), idempotency keys, retries,
// tracing. Scheduler stays dumb; jobs stay pure; guards compose.
package guards

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

// SkipOverlap drops a trigger while the previous run still works.
// (robfig also has SkipIfStillRunning built in; this shows the mechanism.)
func SkipOverlap(job func(ctx context.Context) error) func(ctx context.Context) error {
	var running atomic.Bool

	slog.Info("guard: SkipOverlap wraps job initialized")

	return func(ctx context.Context) error {
		if !running.CompareAndSwap(false, true) {
			slog.Info("skip: previous run still working")

			return nil
		}

		defer running.Store(false)

		return job(ctx)
	}
}

// Jitter sleeps a random 0..max before running: N jobs on the same
// schedule spread out instead of stampeding at once (thundering herd
// at 2am). Respects ctx cancel while waiting.
func Jitter(job func(ctx context.Context) error, max time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		//nolint:gosec // demo jitter: randomness spreads load, not security.
		delay := time.Duration(rand.Int64N(int64(max) + 1))

		slog.Info("jitter delay", "wait", delay.Round(time.Millisecond))

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return job(ctx)
		}
	}
}
