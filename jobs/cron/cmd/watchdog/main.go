// Command watchdog is the SLA countdown R&D sandbox: a job that never
// succeeds plus a watcher that pages on silence. Separate main so the
// cron demo stays clean. Usage: go run ./cmd/watchdog
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

	w := scheduler.New(
		// Never succeeds (Demo always cancels): the watcher pages every round.
		scheduler.Job{Name: "flaky", Schedule: "@every 5s", Run: guards.SkipOverlap(jobs.Demo)},
		// Watchdog: SLA 8s on "flaky" — expect breach pages, by design.
		scheduler.Job{Name: "sla-flaky", Schedule: "@every 10s", Run: jobs.SLAWatcher("flaky", 8*time.Second)},
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
