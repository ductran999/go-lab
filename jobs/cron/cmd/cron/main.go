// Command cron runs the demo job on schedule: scheduler ticks,
// guards protect, jobs work. Thin main: wiring only.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-lab/jobs/cron/internal/guards"
	"go-lab/jobs/cron/internal/jobs"
	"go-lab/jobs/cron/internal/logx"
	"go-lab/jobs/cron/internal/scheduler"
)

func main() {
	os.Exit(run())
}

func run() int {
	slog.SetDefault(slog.New(logx.New(os.Stdout)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	all := []scheduler.Job{
		{Name: "demo", Schedule: "@every 5s", Run: guards.SkipOverlap(jobs.Demo)},
		// Combined: spread starts AND drop overlaps. Either alone leaks.
		{
			Name:     "nightly",
			Schedule: "@every 7s",
			Run:      guards.SkipOverlap(guards.Jitter(jobs.Demo, 2*time.Second)),
		},
		// Watchdog: demo never succeeds (always cancelled), so this pages
		// every round — the SLA countdown shape (deadline + alert on miss).
		{
			Name:     "sla-demo",
			Schedule: "@every 10s",
			Run:      jobs.SLAWatcher("demo", 8*time.Second),
		},
	}

	// Leader election lives in cmd/leader (needs PG_DSN): this binary
	// stays dependency-free of databases.
	w := scheduler.New(all...)

	err := w.StartAll(ctx)
	if err != nil {
		slog.Error("failed to start jobs", "error", err.Error())

		return 1
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	slog.Info("shutting down application...")

	cancel()
	w.StopAll()
	slog.Info("application exited")

	return 0
}
