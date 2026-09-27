// Command migrate creates the users table (gorm AutoMigrate).
// Usage: go run ./cmd/migrate (needs make up).
package main

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ductran999/shared-pkg/environ"
)

type User struct {
	ID   uint
	Name string
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "migrate:", err)

	os.Exit(1)
}

func main() {
	dsn := fmt.Sprintf("host=localhost port=%s user=admin password=password123 dbname=gormdb sslmode=disable",
		environ.Get("PORT", "5436"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fail(err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		fail(err)
	}

	fmt.Println("migrated: users")
}
