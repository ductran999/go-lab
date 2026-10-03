// Package jobs holds one func per usecase: cleanup, digest, sync, warmup.
// Each takes ctx (timeout/cancel honored) and returns error for retry
// decisions by guards. Pure logic, no scheduling inside.
package jobs

import (
	"context"
	"log/slog"
	"time"
)

// Demo simulates 5s work against a 2s timeout: it always cancels.
// Replace with real usecases (same signature).
func Demo(jobCtx context.Context) error {
	slog.Info("working hard...")

	select {
	case <-time.After(5 * time.Second):
		slog.Info("work finished naturally")

		return nil

	case <-jobCtx.Done():
		return jobCtx.Err()
	}
}
