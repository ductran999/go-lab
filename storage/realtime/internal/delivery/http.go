// Package delivery holds the HTTP surface: a token endpoint, a JSON
// writer endpoint, and an authenticated per-tenant SSE stream with
// Last-Event-ID resume.
package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-lab/storage/realtime/internal/auth"
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

// RegisterRoutes mounts POST /token, POST /events and GET /stream.
// Writers and streams require a tenant token; EventSource streams
// pass it as ?token= (no custom headers there).
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/token", h.token)
	protected := r.Group("/", auth.RequireTenant())
	protected.POST("/events", h.create)
	protected.GET("/stream", h.stream)
}

// tokenRequest is the POST /token body. Demo issuance only.
type tokenRequest struct {
	TenantID int `json:"tenant_id"`
}

func (h *Handler) token(c *gin.Context) {
	var req tokenRequest

	err := c.ShouldBindJSON(&req)
	if err != nil || req.TenantID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid tenant_id is required"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": auth.Mint(req.TenantID)})
}

// createRequest is the POST /events body. Tenant comes from the token.
type createRequest struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

func (h *Handler) create(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")

	var req createRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})

		return
	}

	if req.Kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})

		return
	}

	event := events.Event{TenantID: tenantID.(int), Kind: req.Kind, Payload: req.Payload}

	err = h.db.WithContext(c.Request.Context()).Create(&event).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store event"})

		return
	}

	c.JSON(http.StatusCreated, event)
}

// idProbe extracts the event id from a NOTIFY payload (row_to_json).
type idProbe struct {
	ID int `json:"id"`
}

func writeEvent(w http.ResponseWriter, id int, data []byte, flusher http.Flusher) {
	_, _ = fmt.Fprintf(w, "id: %d\ndata: %s\n\n", id, data)

	flusher.Flush()
}

// stream serves Server-Sent Events for the token tenant. It replays
// missed rows after Last-Event-ID, then follows the live Hub.
// A comment heartbeat every 15s keeps intermediaries from idling out.
func (h *Handler) stream(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")

	lastID := 0

	if raw := c.GetHeader("Last-Event-ID"); raw != "" {
		n, convErr := strconv.Atoi(raw)
		if convErr == nil && n > 0 {
			lastID = n
		}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Subscribe before replaying: anything arriving mid-replay stays
	// buffered and is deduplicated by id below.
	ch, unsubscribe := h.hub.Subscribe(tenantID.(int))
	defer unsubscribe()

	var missed []events.Event

	err := h.db.WithContext(c.Request.Context()).
		Where("tenant_id = ? AND id > ?", tenantID, lastID).
		Order("id ASC").Find(&missed).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replay"})

		return
	}

	maxSent := lastID

	for _, e := range missed {
		raw, _ := json.Marshal(e)
		writeEvent(c.Writer, e.ID, raw, c.Writer)
		maxSent = e.ID
	}

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

			var probe idProbe

			err := json.Unmarshal(msg, &probe)
			if err != nil || probe.ID <= maxSent {
				continue
			}

			writeEvent(c.Writer, probe.ID, msg, c.Writer)

			maxSent = probe.ID
		}
	}
}
