// Command bench batch-inserts N rows per PK shape and reports wall
// time plus index bytes (pg_relation_size of the pkey index).
// Usage: go run ./cmd/bench -n 50000 (needs make migrate).
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "bench:", err)

	os.Exit(1)
}

func main() {
	n := flag.Int("n", 50000, "rows per table")

	flag.Parse()

	dsn := fmt.Sprintf("host=localhost port=%s user=admin password=password123 dbname=pkdb sslmode=disable",
		environ.Get("PORT", "5438"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		fail(err)
	}

	benchInt(db, *n)
	benchU4(db, *n)
	benchU7(db, *n)
	benchSizes(db)
	benchPages(db)
}

// benchPages times deep OFFSET pagination per shape: ordered keys scan
// forward, scattered keys sort. Same page (OFFSET 40000, LIMIT 20).
func benchPages(db *gorm.DB) {
	for _, table := range []string{"t_int", "t_u4", "t_u7"} {
		var rows []map[string]any

		start := time.Now()

		err := db.Table(table).Order("id").Offset(40000).Limit(20).Find(&rows).Error
		if err != nil {
			fail(err)
		}

		fmt.Printf("page  %s: %s (%d rows)\n", table, time.Since(start).Round(time.Microsecond), len(rows))
	}
}

func benchInt(db *gorm.DB, n int) {
	rows := make([]map[string]any, 0, n)
	for range n {
		rows = append(rows, map[string]any{"body": "x"})
	}

	start := time.Now()

	err := db.Table("t_int").CreateInBatches(rows, 1000).Error
	if err != nil {
		fail(err)
	}

	fmt.Printf("t_int : %s for %d rows\n", time.Since(start).Round(time.Millisecond), n)
}

func benchU4(db *gorm.DB, n int) {
	rows := make([]map[string]any, 0, n)
	for range n {
		rows = append(rows, map[string]any{"body": "x"})
	}

	start := time.Now()

	err := db.Table("t_u4").CreateInBatches(rows, 1000).Error
	if err != nil {
		fail(err)
	}

	fmt.Printf("t_u4  : %s for %d rows\n", time.Since(start).Round(time.Millisecond), n)
}

func benchU7(db *gorm.DB, n int) {
	rows := make([]map[string]any, 0, n)
	for range n {
		id, err := uuid.NewV7()
		if err != nil {
			fail(err)
		}

		rows = append(rows, map[string]any{"id": id.String(), "body": "x"})
	}

	start := time.Now()

	err := db.Table("t_u7").CreateInBatches(rows, 1000).Error
	if err != nil {
		fail(err)
	}

	fmt.Printf("t_u7  : %s for %d rows\n", time.Since(start).Round(time.Millisecond), n)
}

func benchSizes(db *gorm.DB) {
	var rows []struct {
		Table string
		Bytes int64
	}

	err := db.Raw(`
		SELECT relname AS table, pg_relation_size(indexrelid) AS bytes
		FROM pg_index JOIN pg_class ON pg_class.oid = indexrelid
		WHERE relname IN ('t_int_pkey', 't_u4_pkey', 't_u7_pkey')
		ORDER BY relname`).Scan(&rows).Error
	if err != nil {
		fail(err)
	}

	for _, r := range rows {
		fmt.Printf("index %-10s %d bytes\n", r.Table, r.Bytes)
	}
}
