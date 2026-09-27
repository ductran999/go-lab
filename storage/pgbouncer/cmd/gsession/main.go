// Command gsession shows what a DB session is through GORM: state
// (SET vars, temp tables) sticks to ONE backend process, not to the
// *gorm.DB handle. Pool checkouts land anywhere; Conn() pins one.
// Usage: go run ./cmd/gsession -dsn 'postgres://...:5435/pooldb?sslmode=disable'
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ErrNoDSN = errors.New("missing -dsn")

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gsession:", err)

	os.Exit(1)
}

func show(db *gorm.DB, who string) {
	var value string

	err := db.Raw("SHOW myapp.who").Scan(&value).Error
	if err != nil {
		fmt.Printf("[%s] myapp.who=<unset> (%v)\n", who, err)

		return
	}

	fmt.Printf("[%s] myapp.who=%q\n", who, value)
}

func main() {
	dsn := flag.String("dsn", "", "postgres DSN (required)")

	flag.Parse()

	if *dsn == "" {
		fail(ErrNoDSN)
	}

	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{})
	if err != nil {
		fail(err)
	}

	// 1. Pool checkout: SET may land anywhere; next checkout elsewhere.
	setErr := db.Exec("SET myapp.who = 'pool'").Error
	if setErr != nil {
		fail(setErr)
	}

	fmt.Println("== pool checkouts (maybe different backends) ==")
	show(db, "pool-1")
	show(db, "pool-2")

	// 2. Pinned conn: same backend, state sticks.
	sqlDB, err := db.DB()
	if err != nil {
		fail(err)
	}

	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	fmt.Println("== pinned conn (one backend) ==")

	_, err = conn.ExecContext(context.Background(), "SET myapp.who = 'pinned'")
	if err != nil {
		fail(err)
	}

	var value string

	err = conn.QueryRowContext(context.Background(), "SHOW myapp.who").Scan(&value)
	if err != nil {
		fail(err)
	}

	fmt.Printf("[pinned] myapp.who=%q (sticks: same session)\n", value)
}
