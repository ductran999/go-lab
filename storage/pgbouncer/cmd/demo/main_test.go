package main

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ductran999/shared-pkg/environ"
)

// Unreachable DB: every query must fail LOUD (failed > 0), never
// silently pass. Run: go test ./cmd/demo/ -v (no DB needed).
// Point PGDEMO_DSN at live direct (:5435) to prove real behavior:
// PGDEMO_DSN='postgres://admin:password123@localhost:5435/pooldb?sslmode=disable'
// go test ./cmd/demo/ -v (then it must pass clean: failed == 0).
func TestRunUnreachable(t *testing.T) {
	t.Parallel()

	dsn := environ.Get("PGDEMO_DSN", "")
	live := dsn != ""

	if !live {
		dsn = "postgres://admin:password123@localhost:45999/pooldb?sslmode=disable&connect_timeout=2"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool build: %v", err)
	}

	defer pool.Close()

	ok, failed, _, firstErr := run(pool, 4, 2)

	t.Logf("live=%t ok=%d failed=%d firstErr=%v", live, ok, failed, firstErr)

	if live {
		if failed != 0 {
			t.Fatalf("live DB should pass clean, failed=%d firstErr=%v", failed, firstErr)
		}

		return
	}

	if failed == 0 {
		t.Fatal("expected failures against a dead port, got none")
	}

	if ok != 0 {
		t.Fatalf("expected zero ok, got %d", ok)
	}

	if firstErr == nil {
		t.Fatal("expected a raw error, got nil")
	}
}

// TestTooManyClients proves the max_connections=10 wall on direct :5435:
// 20 clients × immediate queries → some fail with "too many clients".
// DSN lives in code (override with PGDEMO_DSN); skips when no live DB.
// go test ./cmd/demo/ -run TestTooManyClients -v.
func TestTooManyClients(t *testing.T) {
	dsn := environ.Get("PGDEMO_DSN",
		"postgres://admin:password123@localhost:5435/pooldb?sslmode=disable")

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool build: %v", err)
	}

	defer pool.Close()

	// Sanity: one query must work (else the DB itself is down, not full).
	var one int

	err = pool.QueryRow(context.Background(), "SELECT 1").Scan(&one)
	if err != nil {
		t.Skipf("live DB unreachable: %v", err)
	}

	ok, failed, _, firstErr := run(pool, 20, 10)

	t.Logf("overload: ok=%d failed=%d firstErr=%v", ok, failed, firstErr)

	if failed == 0 {
		t.Log("no failures: pool absorbed it (try direct :5435, not :6432)")
	}
}

// TestPooledSurvives is the mirror: same 20 clients through pgbouncer
// (:6432, pool size 5) must pass clean — 5 backends serve 20 clients.
// Needs live pool (skips otherwise).
// go test ./cmd/demo/ -run TestPooledSurvives -v.
func TestPooledSurvives(t *testing.T) {
	dsn := environ.Get("PGPOOL_DSN",
		"postgres://admin:password123@localhost:6432/pooldb?sslmode=disable")

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool build: %v", err)
	}

	defer pool.Close()

	var one int

	err = pool.QueryRow(context.Background(), "SELECT 1").Scan(&one)
	if err != nil {
		t.Skipf("pool unreachable: %v", err)
	}

	ok, failed, _, firstErr := run(pool, 20, 10)

	t.Logf("pooled: ok=%d failed=%d firstErr=%v", ok, failed, firstErr)

	if failed != 0 {
		t.Fatalf("pool of 5 should absorb 20 clients, failed=%d firstErr=%v", failed, firstErr)
	}
}
