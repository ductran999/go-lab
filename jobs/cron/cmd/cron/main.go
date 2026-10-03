// Command cron runs the demo job on schedule: scheduler ticks,
// guards protect, jobs work. Thin main: wiring only.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go-lab/jobs/cron/internal/guards"
	"go-lab/jobs/cron/internal/jobs"
	"go-lab/jobs/cron/internal/scheduler"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := scheduler.New(
		scheduler.Job{Name: "demo", Schedule: "@every 3s", Run: guards.SkipOverlap(jobs.Demo)},
	)

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
