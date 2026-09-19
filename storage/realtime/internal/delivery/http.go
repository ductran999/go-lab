// Package delivery holds the HTTP surface: a JSON writer endpoint
// and a per-tenant SSE stream fed by the Hub.
package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-lab/storage/realtime/internal/bus"
	"go-lab/storage/realtime/internal/events"
)

// Handler writes events and streams them per tenant.
type Handler struct {
	db  *gorm.DB
	hub *bus.Hub
}

// NewHandler builds a Handler. It panics on nil dependencies.
func NewHandler(db *gorm.DB, hub *bus.Hub) *Handler {
	if db == nil {
		panic("delivery: nil gorm.DB")
	}

	if hub == nil {
		panic("delivery: nil bus.Hub")
	}

	return &Handler{db: db, hub: hub}
}

// RegisterRoutes mounts POST /events and GET /stream.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/events", h.create)
	r.GET("/stream", h.stream)
}

// createRequest is the POST /events body.
type createRequest struct {
	TenantID int             `json:"tenant_id"`
	Kind     string          `json:"kind"`
	Payload  json.RawMessage `json:"payload"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})

		return
	}

	if req.TenantID <= 0 || req.Kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id and kind are required"})

		return
	}

	event := events.Event{TenantID: req.TenantID, Kind: req.Kind, Payload: req.Payload}

	err = h.db.WithContext(c.Request.Context()).Create(&event).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store event"})

		return
	}

	c.JSON(http.StatusCreated, event)
}

// stream serves Server-Sent Events for one tenant (?tenant_id=N).
// A comment heartbeat every 15s keeps intermediaries from idling out.
func (h *Handler) stream(c *gin.Context) {
	tenantID, err := strconv.Atoi(c.Query("tenant_id"))
	if err != nil || tenantID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid ?tenant_id= is required"})

		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := h.hub.Subscribe(tenantID)
	defer unsubscribe()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			_, _ = fmt.Fprint(c.Writer, ": ping\n\n")
			c.Writer.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}

			_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			c.Writer.Flush()
		}
	}
}
