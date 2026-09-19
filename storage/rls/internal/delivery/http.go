// Package delivery holds HTTP presentation: DTOs shared by every
// entry point (CLI demos and the gin server) plus the gin handler.
// JSON tags live here, never on domain entities.
package delivery

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-lab/storage/rls/internal/domain"
	"go-lab/storage/rls/internal/usecase"
)

// DocumentHandler serves tenant documents over HTTP.
type DocumentHandler struct {
	reporter *usecase.Reporter
	secret   string
}

// NewDocumentHandler builds a handler over any repository.
// Tenant comes from the URL path (explicit scoping demo).
// It panics on nil: a missing dependency must fail at wiring time.
func NewDocumentHandler(repo domain.DocumentRepository) *DocumentHandler {
	return &DocumentHandler{reporter: usecase.NewReporter(repo)}
}

// NewRLSDocumentHandler builds a handler whose tenant comes from JWT
// via RequireTenant middleware, never from URL params.
func NewRLSDocumentHandler(repo domain.DocumentRepository, secret string) *DocumentHandler {
	return &DocumentHandler{reporter: usecase.NewReporter(repo), secret: secret}
}

// RegisterRoutes mounts GET /tenants/:id/docs on the group.
func (h *DocumentHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/tenants/:id/docs", h.listByTenant)
}

// RegisterRLSRoutes mounts GET /docs behind JWT auth on the group.
func (h *DocumentHandler) RegisterRLSRoutes(g *gin.RouterGroup) {
	g.Use(RequireTenant(h.secret))
	g.GET("/docs", h.listMine)
}

func (h *DocumentHandler) listMine(c *gin.Context) {
	tenantID, ok := tenantFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})

		return
	}

	report, err := h.reporter.BuildReport(c.Request.Context(), []int{tenantID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list documents"})

		return
	}

	c.JSON(http.StatusOK, ToDTOs(report[0].Docs))
}

func (h *DocumentHandler) listByTenant(c *gin.Context) {
	tenantID, err := strconv.Atoi(c.Param("id"))
	if err != nil || tenantID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})

		return
	}

	report, err := h.reporter.BuildReport(c.Request.Context(), []int{tenantID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list documents"})

		return
	}

	c.JSON(http.StatusOK, ToDTOs(report[0].Docs))
}
