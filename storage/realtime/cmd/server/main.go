// Command server bridges Postgres LISTEN/NOTIFY to browser SSE:
// POST /events inserts a row (trigger NOTIFYs), GET /stream fans out
// per tenant (?tenant_id=N). A dedicated pgx connection LISTENs;
// dbkit handles writes.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/ductran999/dbkit"
	"github.com/ductran999/shared-pkg/environ"

	"go-lab/storage/realtime/internal/bus"
	"go-lab/storage/realtime/internal/delivery"
	"go-lab/storage/realtime/internal/listener"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func dsn() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		environ.Get("DB_HOST", "localhost"),
		environ.GetInt("DB_PORT", 5434),
		environ.Get("DB_USER", "admin"),
		environ.Get("DB_PASSWORD", "password123"),
		environ.Get("DB_NAME", "realtime"),
	)
}

func pgConfig() dbkit.PostgreSQLConfig {
	return dbkit.PostgreSQLConfig{
		Config: dbkit.Config{
			Host:     environ.Get("DB_HOST", "localhost"),
			Port:     environ.GetInt("DB_PORT", 5434),
			Username: environ.Get("DB_USER", "admin"),
			Password: environ.Get("DB_PASSWORD", "password123"),
			Database: environ.Get("DB_NAME", "realtime"),
		},
		SSLMode: dbkit.PgSSLDisable,
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	level := slog.LevelInfo
	if os.Getenv("LOG_DEBUG") != "" {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	conn, err := dbkit.NewPostgreSQLConnection(pgConfig())
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	hub := bus.NewHub()

	go func() {
		err := listener.Start(ctx, dsn(), "events", hub)
		if err != nil && ctx.Err() == nil {
			slog.Error("listener stopped", "error", err)
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	// SSE streams only end on client disconnect: their "latency" is
	// connection lifetime, not slowness. Skip them from access logs.
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/stream"}}))

	// Demo frontend. Served from ./web, so run from storage/realtime
	// (e.g. via make run-server) for the path to resolve.
	r.StaticFile("/", "./web/index.html")

	delivery.NewHandler(conn.DB(), hub).RegisterRoutes(r)

	addr := ":" + environ.Get("PORT", "8082")

	slog.Info("serving realtime API", "addr", addr)

	go func() {
		err := r.Run(addr)
		if err != nil {
			slog.Error("http stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
}
