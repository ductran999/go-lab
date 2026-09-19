// Command server exposes both repository implementations over HTTP:
// GET /scoped/tenants/:id/docs (explicit tenant param) and
// GET /rls/docs (tenant from JWT, never in params).
// Same results, different enforcement points. Compare with curl.
package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/ductran999/shared-pkg/environ"

	"go-lab/storage/rls/internal/bootstrap"
	"go-lab/storage/rls/internal/delivery"
	"go-lab/storage/rls/internal/infrastructure/rls"
	"go-lab/storage/rls/internal/infrastructure/scoped"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func main() {
	conn, err := bootstrap.OpenPostgres()
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	r := gin.Default()

	delivery.NewDocumentHandler(scoped.NewDocumentRepository(conn.DB)).RegisterRoutes(r.Group("/scoped"))
	delivery.NewRLSDocumentHandler(
		rls.NewDocumentRepository(conn.DB),
		environ.Get("RLS_JWT_SECRET", "dev-secret-change-me-please"),
	).RegisterRLSRoutes(r.Group("/rls"))

	addr := ":" + environ.Get("PORT", "8081")

	slog.Info("serving comparison API", "addr", addr)

	err = r.Run(addr)
	if err != nil {
		fail(err)
	}
}
