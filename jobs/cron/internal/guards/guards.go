// Package guards wraps job funcs with cross-cutting concerns:
// overlap locks (SkipIfStillRunning), idempotency keys, retries,
// tracing. Scheduler stays dumb; jobs stay pure; guards compose.
package guards

import (
	"context"
	"log/slog"
	"sync/atomic"
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
