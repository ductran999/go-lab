// Command demo hammers Postgres with N concurrent app clients, Q
// queries each, and reports wall time + errors. Point DSN at :5435
// (direct, max 10) vs :6432 (pooled) to feel the difference.
// Usage: go run ./cmd/demo -dsn 'postgres://admin:pw@localhost:6432/pooldb?sslmode=disable' -clients 50 -q 20
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoDSN = errors.New("missing -dsn")

func fail(err error) {
	fmt.Fprintln(os.Stderr, "demo:", err)

	os.Exit(1)
}

func main() {
	dsn := flag.String("dsn", "", "postgres DSN (required)")
	clients := flag.Int("clients", 50, "concurrent app clients")
	queries := flag.Int("q", 20, "queries per client")

	flag.Parse()

	if *dsn == "" {
		fail(ErrNoDSN)
	}

	pool, err := pgxpool.New(context.Background(), *dsn)
	if err != nil {
		fail(err)
	}
	defer pool.Close()

	ok, failed, el, firstErr := run(pool, *clients, *queries)
	if firstErr != nil {
		fmt.Fprintf(os.Stderr, "demo: first error: %v\n", firstErr)
	}

	fmt.Printf("clients=%d q=%d ok=%d failed=%d wall=%s qps=%.0f\n",
		*clients, *queries, ok, failed, el.Round(time.Millisecond),
		float64(ok)/el.Seconds())
}

// run fires clients×queries SELECTs, counting ok vs failed.
// Errors are counted (and logged by callers), never swallowed.
func run(pool *pgxpool.Pool, clients, queries int) (int64, int64, time.Duration, error) {
	var okCount, failedCount atomic.Int64

	var firstErr atomic.Value

	var wg sync.WaitGroup

	start := time.Now()

	for range clients {
		wg.Go(func() {
			for range queries {
				var n int

				err := pool.QueryRow(context.Background(), "SELECT 1").Scan(&n)
				if err != nil {
					failedCount.Add(1)
					firstErr.CompareAndSwap(nil, err)

					return
				}

				okCount.Add(1)
			}
		})
	}

	wg.Wait()

	var fe error

	if v := firstErr.Load(); v != nil {
		fe, _ = v.(error)
	}

	return okCount.Load(), failedCount.Load(), time.Since(start), fe
}
