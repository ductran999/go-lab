// Command demo shows what gorm.Session flags change, using DryRun:
// SQL is generated and printed, never executed. gorm.Open still dials
// once, so point at this lab's own DB (:5436) — nothing is written.
package main

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "demo:", err)

	os.Exit(1)
}

type User struct {
	ID   uint
	Name string
}

var fired []string

// BeforeCreate is a MODEL hook: SkipHooks skips it. (Processor callbacks
// registered via Callback().Register() are different: they build SQL and
// always run, even in DryRun — SkipHooks never touches them.)
func (u *User) BeforeCreate(tx *gorm.DB) error {
	fired = append(fired, "before")

	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	fired = append(fired, "after")

	return nil
}

func show(title string, db *gorm.DB) {
	stmt := db.Session(&gorm.Session{DryRun: true}).Find(&[]User{}).Statement

	fmt.Printf("== %s ==\nSQL: %v\n\n", title, stmt.SQL.String())
}

func main() {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=localhost port=5436 user=admin password=password123 dbname=gormdb sslmode=disable",
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		fail(err)
	}

	// Hooks proof: model hooks fire on normal runs (DryRun still runs
	// callbacks — only SQL exec is skipped), SkipHooks silences them.
	// NOTE: processor callbacks (Callback().Register) always run;
	// SkipHooks covers model hooks only (see source: callbacks/create.go).

	// 1. Default: writes wrap in a transaction (BEGIN/COMMIT in SQL log).
	tx := db.Session(&gorm.Session{DryRun: true}).Create(&User{Name: "u1"})
	fmt.Printf("== default (tx wrapped) ==\nSQL: %v\nVars: %v\n\n", tx.Statement.SQL.String(), tx.Statement.Vars)

	// 2. SkipDefaultTransaction: bare INSERT, no explicit BEGIN/COMMIT.
	// (Postgres still wraps it in an implicit single-statement tx:
	// atomic either way. The flag only drops the explicit wrapper —
	// it matters for multi-statement units, not atomicity.)
	bare := db.Session(&gorm.Session{DryRun: true, SkipDefaultTransaction: true}).Create(&User{Name: "u2"})
	fmt.Printf("== SkipDefaultTransaction ==\nSQL: %v\n\n", bare.Statement.SQL.String())

	// 3. SkipHooks: with hooks registered above, normal run fires both...
	fired = nil
	hooked := db.Session(&gorm.Session{DryRun: true}).Create(&User{Name: "u3"})
	fmt.Printf("== hooks on ==\nSQL: %v\nfired: %v\n\n", hooked.Statement.SQL.String(), fired)

	// ...SkipHooks leaves the log empty. Proof, not promise.
	fired = nil
	skipped := db.Session(&gorm.Session{DryRun: true, SkipHooks: true}).Create(&User{Name: "u3"})
	fmt.Printf("== SkipHooks ==\nSQL: %v\nfired: %v\n\n", skipped.Statement.SQL.String(), fired)

	// 4. AllowGlobalUpdate guard: bare Update without Where errors.
	unguarded := db.Session(&gorm.Session{DryRun: true}).Model(&User{}).Update("name", "x")
	fmt.Printf("== global update blocked ==\nError: %v\n\n", unguarded.Error)

	show("plain Find", db)
}
