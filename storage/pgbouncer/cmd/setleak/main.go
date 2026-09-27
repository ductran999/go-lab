// Command setleak proves SET (session-level) contaminates pooled
// connections while SET LOCAL does not. Pool max 1 conn forces the
// same server backend twice: backend PIDs prove it, values prove the leak.
// Usage: go run ./cmd/setleak -dsn 'postgres://...:6432/pooldb?sslmode=disable'
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
	fmt.Fprintln(os.Stderr, "setleak:", err)

	os.Exit(1)
}

// backendOf returns the server PID behind a checkout.
func backendOf(ctx context.Context, conn *pgxpool.Conn) int32 {
	var pid int32

	err := conn.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid)
	if err != nil {
		fail(err)
	}

	return pid
}

// localPort returns the app-side socket of a checkout (hop 1: app).
func localPort(conn *pgxpool.Conn) string {
	return conn.Conn().PgConn().Conn().LocalAddr().String()
}

// bouncerServers asks the admin console how many server backends the
// pool holds right now (hop 2: bouncer). Needs ADMIN_USERS set.
func bouncerServers(ctx context.Context, dsn string) string {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return "(bad dsn)"
	}

	cfg.ConnConfig.Database = "pgbouncer"

	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return "(admin refused)"
	}

	defer admin.Close()

	rows, err := admin.Query(ctx, "SHOW SERVERS")
	if err != nil {
		return "(admin console refused: set ADMIN_USERS)"
	}

	defer rows.Close()

	count := 0

	for rows.Next() {
		count++
	}

	return fmt.Sprintf("%d server backend(s) held", count)
}

// showVar reads a custom var, reporting missing as <unset>.
func showVar(ctx context.Context, conn *pgxpool.Conn, name string) string {
	var value string

	err := conn.QueryRow(ctx, "SHOW "+name).Scan(&value)
	if err != nil {
		return "<unset>"
	}

	return value
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

	// Force one app conn so both checkouts land on one server backend.
	pool.Config().MaxConns = 1

	ctx := context.Background()

	fmt.Println("== 1. plain SET (session-level) ==")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	fmt.Printf("[checkout A] app=%s backend pid=%d, myapp.tenant=%s\n",
		localPort(conn), backendOf(ctx, conn), showVar(ctx, conn, "myapp.tenant"))

	_, err = conn.Exec(ctx, "SET myapp.tenant = 't1'")
	if err != nil {
		fail(err)
	}

	fmt.Printf("[checkout A] SET myapp.tenant='t1' (app=%s backend pid=%d)\n",
		localPort(conn), backendOf(ctx, conn))
	conn.Release()
	fmt.Println("[checkout A] released back to pool")
	fmt.Println("[bouncer]", bouncerServers(ctx, *dsn))

	conn, err = pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	fmt.Printf("[checkout B] app=%s backend pid=%d, myapp.tenant=%s\n",
		localPort(conn), backendOf(ctx, conn), showVar(ctx, conn, "myapp.tenant"))
	conn.Release()

	fmt.Println("== 2. SET LOCAL (transaction-scoped) ==")

	conn, err = pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	fmt.Printf("[checkout C] backend pid=%d\n", backendOf(ctx, conn))

	tx, err := conn.Begin(ctx)
	if err != nil {
		fail(err)
	}

	_, err = tx.Exec(ctx, "SET LOCAL myapp.tenant = 't2'")
	if err != nil {
		fail(err)
	}

	var inside string

	err = tx.QueryRow(ctx, "SHOW myapp.tenant").Scan(&inside)
	if err != nil {
		fail(err)
	}

	fmt.Printf("[tx] SET LOCAL myapp.tenant='t2', inside tx reads=%q\n", inside)

	err = tx.Commit(ctx)
	if err != nil {
		fail(err)
	}

	fmt.Println("[tx] COMMITTED")

	conn.Release()
	fmt.Println("[checkout C] released back to pool")

	conn, err = pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	defer conn.Release()

	fmt.Printf("[checkout D] app=%s backend pid=%d, myapp.tenant=%s\n",
		localPort(conn), backendOf(ctx, conn), showVar(ctx, conn, "myapp.tenant"))
}
