package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go-lab/storage/rls/internal/domain"
)

var errDBDown = errors.New("db down")

// stubRepository is a hand-written fake: narrow interfaces need
// no mock framework, a small struct is clearer and cheaper.
type stubRepository struct {
	byTenant map[int][]domain.Document
	err      error
}

func (s stubRepository) ListByTenant(_ context.Context, tenantID int) ([]domain.Document, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.byTenant[tenantID], nil
}

func TestBuildReportGroupsTenantsInOrder(t *testing.T) {
	repo := stubRepository{byTenant: map[int][]domain.Document{
		1: {{ID: 1, TenantID: 1, Body: "a"}, {ID: 2, TenantID: 1, Body: "b"}},
		2: {{ID: 3, TenantID: 2, Body: "c"}},
	}}

	reporter := NewReporter(repo)

	report, err := reporter.BuildReport(context.Background(), []int{1, 2})
	require.NoError(t, err)

	require.Len(t, report, 2)
	require.Equal(t, 1, report[0].TenantID)
	require.Len(t, report[0].Docs, 2)
	require.Equal(t, 2, report[1].TenantID)
	require.Len(t, report[1].Docs, 1)
}

func TestBuildReportPropagatesRepoError(t *testing.T) {
	repo := stubRepository{err: errDBDown}

	reporter := NewReporter(repo)

	report, err := reporter.BuildReport(context.Background(), []int{1})
	require.Error(t, err)
	require.Nil(t, report)
}

func TestBuildReportEmptyTenants(t *testing.T) {
	reporter := NewReporter(stubRepository{})

	report, err := reporter.BuildReport(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, report)
}

func TestNewReporterPanicsOnNil(t *testing.T) {
	require.Panics(t, func() { NewReporter(nil) })
}
