// Command migrate creates the three PK-shape tables.
package main

import (
	"fmt"
	"os"

	"github.com/ductran999/shared-pkg/environ"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TInt struct {
	ID   uint `gorm:"primaryKey;autoIncrement"`
	Body string
}

type TU4 struct {
	ID   string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Body string
}

type TU7 struct {
	ID   string `gorm:"type:uuid;primaryKey"`
	Body string
}

func (TInt) TableName() string { return "t_int" }

func (TU4) TableName() string { return "t_u4" }

func (TU7) TableName() string { return "t_u7" }

func fail(err error) {
	fmt.Fprintln(os.Stderr, "migrate:", err)

	os.Exit(1)
}

func main() {
	dsn := fmt.Sprintf("host=localhost port=%s user=admin password=password123 dbname=pkdb sslmode=disable",
		environ.Get("PORT", "5438"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fail(err)
	}

	err = db.AutoMigrate(&TInt{}, &TU4{}, &TU7{})
	if err != nil {
		fail(err)
	}

	fmt.Println("migrated: t_int, t_u4, t_u7")
}
