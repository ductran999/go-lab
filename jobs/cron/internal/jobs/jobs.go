// Package jobs holds one func per usecase: cleanup, digest, sync, warmup.
// Each takes ctx (timeout/cancel honored) and returns error for retry
// decisions by guards. Pure logic, no scheduling inside.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var ErrSLABreach = errors.New("sla breach: job silent past window")

// Demo simulates 5s work against a 2s timeout: it always cancels.
// Replace with real usecases (same signature).
func Demo(jobCtx context.Context) error {
	slog.Info("working hard...")

	select {
	case <-time.After(5 * time.Second):
		slog.Info("work finished naturally")

		Track.Complete("demo")

		return nil

	case <-jobCtx.Done():
		return jobCtx.Err()
	}
}

// Track is the demo completion registry: jobs mark done, the SLA
// watcher reads. Real systems persist this (DB/Redis), not memory.
var Track = NewTracker()

// Tracker records last-success timestamps per job name.
type Tracker struct {
	mu   sync.Mutex
	done map[string]time.Time
}

// NewTracker builds an empty Tracker.
func NewTracker() *Tracker {
	return &Tracker{done: make(map[string]time.Time)}
}

// Complete marks a job success now.
func (t *Tracker) Complete(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.done[name] = time.Now()
}

// Since returns time since last success (huge if never).
func (t *Tracker) Since(name string) time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	last, ok := t.done[name]
	if !ok {
		return time.Hour * 24 * 365
	}

	return time.Since(last)
}

// SLAWatcher pages (logs) when a job missed its deadline: no success
// within window. The watchdog pattern: deadline + alert on miss.
func SLAWatcher(name string, window time.Duration) func(ctx context.Context) error {
	return func(_ context.Context) error {
		if ago := Track.Since(name); ago > window {
			slog.Info("SLA breach: paging on-call",
				"job", name, "window", window.String(), "last_ok_ago", ago.Round(time.Second).String())

			return fmt.Errorf("%w: %s silent for %s", ErrSLABreach, name, ago.Round(time.Second))
		}

		slog.Info("SLA ok", "job", name)

		return nil
	}
}
