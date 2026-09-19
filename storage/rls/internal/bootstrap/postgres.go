// Package bootstrap wires shared infrastructure for the demo commands:
// a dbkit Postgres connection built from env, with SQL logging attached.
// Each command keeps its own demo flow; only connection setup is shared.
package bootstrap

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ductran999/dbkit"
	"github.com/ductran999/shared-pkg/environ"
)

// Postgres bundles a GORM handle with its closer.
type Postgres struct {
	DB   *gorm.DB
	conn dbkit.Connection
}

// OpenPostgres connects to Postgres using env config (DB_HOST,
// DB_PORT, DB_USER, DB_PASSWORD, DB_NAME with lab defaults).
// Callers must Close when done.
func OpenPostgres() (*Postgres, error) {
	conn, err := dbkit.NewPostgreSQLConnection(pgConfig())
	if err != nil {
		return nil, err
	}

	db := conn.DB()
	db.Logger = logger.Default.LogMode(logger.Info)

	return &Postgres{DB: db, conn: conn}, nil
}

// Close releases the underlying connection.
func (p *Postgres) Close() error {
	return p.conn.Close()
}

func pgConfig() dbkit.PostgreSQLConfig {
	return dbkit.PostgreSQLConfig{
		Config: dbkit.Config{
			Host:     environ.Get("DB_HOST", "localhost"),
			Port:     environ.GetInt("DB_PORT", 5433),
			Username: environ.Get("DB_USER", "admin"),
			Password: environ.Get("DB_PASSWORD", "password123"),
			Database: environ.Get("DB_NAME", "rls_lab"),
		},
		SSLMode: dbkit.PgSSLDisable,
	}
}
