package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"go-lab/storage/rls/internal/domain"
)

// errDBDown simulates repository failure.
var errDBDown = errors.New("db down")

// stubRepository fakes tenant visibility without a database.
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

func serve(t *testing.T, repo domain.DocumentRepository, target string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	NewDocumentHandler(repo).RegisterRoutes(r.Group("/test"))

	req := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestListByTenantReturnsArray(t *testing.T) {
	repo := stubRepository{byTenant: map[int][]domain.Document{
		1: {{ID: 1, TenantID: 1, Body: "a"}},
	}}

	w := serve(t, repo, "/test/tenants/1/docs")
	require.Equal(t, http.StatusOK, w.Code)

	var body []DocumentDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, "a", body[0].Body)
}

func TestListByTenantEmptyEncodesArrayNotNull(t *testing.T) {
	w := serve(t, stubRepository{}, "/test/tenants/99/docs")
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `[]`, w.Body.String())
}

func TestListByTenantRejectsInvalidTenant(t *testing.T) {
	for _, target := range []string{"/test/tenants/0/docs", "/test/tenants/-1/docs", "/test/tenants/abc/docs"} {
		w := serve(t, stubRepository{}, target)
		require.Equal(t, http.StatusBadRequest, w.Code, target)
	}
}

func TestListByTenantRepoErrorIs500(t *testing.T) {
	repo := stubRepository{err: errDBDown}

	w := serve(t, repo, "/test/tenants/1/docs")
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestToDTOsNeverNil(t *testing.T) {
	require.NotNil(t, ToDTOs(nil))

	dtos := ToDTOs([]domain.Document{{ID: 1, TenantID: 1, Body: "a"}})
	require.Len(t, dtos, 1)
	require.Equal(t, 1, dtos[0].TenantID)
}
