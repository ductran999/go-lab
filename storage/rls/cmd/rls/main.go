// Command rls runs the tenant-isolation demo with Row Level Security:
// queries carry no tenant filter, the policy decides visibility from
// transaction-local context. Compare with the scoped command, which
// enforces the same isolation with an explicit WHERE.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ductran999/shared-pkg/pretty/display"

	"go-lab/storage/rls/internal/bootstrap"
	"go-lab/storage/rls/internal/delivery"
	"go-lab/storage/rls/internal/infrastructure/rls"
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

	repo := rls.NewDocumentRepository(conn.DB)
	reporter := usecase.NewReporter(repo)

	slog.Info("starting RLS demo (policy filtering, no WHERE)")

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
