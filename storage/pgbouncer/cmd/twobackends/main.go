// Command twobackends holds two OPEN transactions at once and prints
// both server PIDs: different PIDs prove distinct backends serve one
// pool. Note: held checkouts alone share one backend (transaction mode
// assigns backends per TRANSACTION, not per checkout) — the txs must
// stay open. Usage: go run ./cmd/twobackends -dsn '...:6432/...'
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoDSN = errors.New("missing -dsn")

func fail(err error) {
	fmt.Fprintln(os.Stderr, "twobackends:", err)

	os.Exit(1)
}

func main() {
	dsn := flag.String("dsn", "", "pgbouncer DSN (required)")

	flag.Parse()

	if *dsn == "" {
		fail(ErrNoDSN)
	}

	pool, err := pgxpool.New(context.Background(), *dsn)
	if err != nil {
		fail(err)
	}

	defer pool.Close()

	ctx := context.Background()

	// Hold two OPEN transactions: only then does the pool hand out two
	// backends at once (idle checkouts share one).
	first, err := pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	defer first.Release()

	tx1, err := first.Begin(ctx)
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = tx1.Rollback(ctx)
	}()

	second, err := pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	defer second.Release()

	tx2, err := second.Begin(ctx)
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = tx2.Rollback(ctx)
	}()

	var pid1, pid2 int32

	err = tx1.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid1)
	if err != nil {
		fail(err)
	}

	err = tx2.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid2)
	if err != nil {
		fail(err)
	}

	fmt.Printf("checkout 1 -> backend pid=%d\ncheckout 2 -> backend pid=%d\n", pid1, pid2)

	if pid1 == pid2 {
		fmt.Println("same PID: pool serialized onto one backend (pool size 1?)")
	} else {
		fmt.Println("different PIDs: two backends serve one pool concurrently")
	}
}
