// Command prepdie proves server-side prepared statements die across
// checkouts in transaction mode: PREPARE on one backend, EXECUTE from
// another gets "prepared statement does not exist". Ten rounds show
// it failing most of the time (5 backends, random landing).
// Usage: go run ./cmd/prepdie -dsn 'postgres://...:6432/pooldb?sslmode=disable'
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
	fmt.Fprintln(os.Stderr, "prepdie:", err)

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

	// PREPARE once, on whatever backend this checkout lands.
	first, err := pool.Acquire(ctx)
	if err != nil {
		fail(err)
	}

	_, err = first.Conn().Prepare(ctx, "stmt1", "SELECT $1::int")
	if err != nil {
		fail(err)
	}

	first.Release()
	fmt.Println("PREPARE stmt1 on one backend, released")

	// EXECUTE from fresh checkouts: different backend, statement gone.
	var lived, died int

	for range 10 {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			fail(err)
		}

		var n int

		err = conn.Conn().QueryRow(ctx, "EXECUTE stmt1(1)").Scan(&n)
		if err != nil {
			died++
		} else {
			lived++
		}

		conn.Release()
	}

	fmt.Printf("EXECUTE over 10 checkouts: lived=%d died=%d\n", lived, died)
	fmt.Println("lesson: server-side PREPARE is backend-local;")
	fmt.Println("tx mode + pool = never rely on it (pgx prepares client-side)")
}
