// Command scoped runs the tenant-isolation demo with manual scoping:
// every query carries an explicit WHERE on tenant_id. Compare with
// the rls command, which enforces the same isolation via policy.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ductran999/shared-pkg/pretty/display"

	"go-lab/storage/rls/internal/bootstrap"
	"go-lab/storage/rls/internal/delivery"
	"go-lab/storage/rls/internal/infrastructure/scoped"
	"go-lab/storage/rls/internal/usecase"
)

func fail(err error) {
	slog.Error("demo failed", "error", err)

	os.Exit(1)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := bootstrap.OpenPostgres()
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	repo := scoped.NewDocumentRepository(conn.DB)
	reporter := usecase.NewReporter(repo)

	slog.Info("starting scoped demo (explicit WHERE per query)")

	report, err := reporter.BuildReport(ctx, []int{1, 2})
	if err != nil {
		fail(err)
	}

	for _, td := range report {
		slog.Info("tenant docs", "tenant", td.TenantID, "count", len(td.Docs))

		printErr := display.PrintJSON(delivery.ToDTOs(td.Docs))
		if printErr != nil {
			fail(printErr)
		}
	}
}
