// Package guards wraps job funcs with cross-cutting concerns:
// overlap locks (SkipIfStillRunning), idempotency keys, retries,
// tracing. Scheduler stays dumb; jobs stay pure; guards compose.
package guards

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

// Leader runs the job only while holding a Postgres advisory lock:
// N replicas tick, one wins, rest standby. Lock releases with the
// session (pool checkout held for the whole run).
func Leader(pool *pgxpool.Pool, key int64, job func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}

		defer conn.Release()

		var locked bool

		err = conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked)
		if err != nil {
			return err
		}

		if !locked {
			slog.Info("standby: leader holds the lock")

			return nil
		}

		defer func() {
			_, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", key)
		}()

		slog.Info("leader: lock acquired, running")

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

// LeaderRedis is the same election over Redis: SET key uuid NX EX 30.
// SET NX = set if Not eXists, atomically: first writer wins, rest get
// nil (standby). TTL auto-releases on crash (unlike PG session locks,
// Redis needs explicit expiry). Release checks uuid via Lua so a slow
// ex-leader never deletes the new leader's lock.
func LeaderRedis(
	rdb *redis.Client, key string, ttl time.Duration, job func(ctx context.Context) error,
) func(ctx context.Context) error {
	id := uuid.NewString()

	return func(ctx context.Context) error {
		got, err := rdb.SetArgs(ctx, key, id, redis.SetArgs{Mode: "NX", TTL: ttl}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return err
		}

		if got == "" {
			slog.Info("standby (redis): leader holds the lock")

			return nil
		}

		slog.Info("leader (redis): lock acquired, running")

		defer func() {
			script := redis.NewScript(
				`if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) else return 0 end`)
			_, _ = script.Run(ctx, rdb, []string{key}, id).Result()
		}()

		return job(ctx)
	}
}
