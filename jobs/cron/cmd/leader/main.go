// Command leader runs one guarded job under PG advisory-lock election:
// N replicas tick, one runs, rest standby. Kill the leader to watch
// failover. Usage:
// PG_DSN='postgres://admin:password123@localhost:5435/pooldb?sslmode=disable' go run ./cmd/leader
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"go-lab/jobs/cron/internal/guards"
	"go-lab/jobs/cron/internal/jobs"
	"go-lab/jobs/cron/internal/logx"
	"go-lab/jobs/cron/internal/scheduler"

	"github.com/ductran999/shared-pkg/environ"
)

func main() {
	os.Exit(run())
}

func run() int {
	slog.SetDefault(slog.New(logx.New(os.Stdout)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, environ.Get("PG_DSN", ""))
	if err != nil {
		slog.Error("leader pool failed", "error", err.Error())

		return 1
	}

	defer pool.Close()

	// Arbiter pick: REDIS_ADDR set → Redis SET NX lock;
	// otherwise Postgres advisory lock. Same shape, two referees.
	run := guards.Leader(pool, 4242, jobs.Demo)

	if addr := environ.Get("REDIS_ADDR", ""); addr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: addr})

		run = guards.LeaderRedis(rdb, "leader-demo", 10*time.Second, jobs.Demo)
	}

	w := scheduler.New(
		scheduler.Job{
			Name:     "leader-demo",
			Schedule: "@every 5s",
			Run:      run,
		},
	)

	err = w.StartAll(ctx)
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
