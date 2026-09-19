// Package bus fans Postgres notifications out to SSE subscribers.
// Subscriptions are per tenant; slow consumers drop messages instead
// of blocking the publisher.
package bus

import (
	"log/slog"
	"sync"
)

// Hub routes raw JSON payloads to tenant subscribers.
type Hub struct {
	mu   sync.RWMutex
	subs map[int]map[chan []byte]struct{}
}

// NewHub builds an empty Hub.
func NewHub() *Hub {
	return &Hub{subs: make(map[int]map[chan []byte]struct{})}
}

// Subscribe registers a buffered channel for one tenant.
// Call the returned function to unsubscribe (usually deferred).
func (h *Hub) Subscribe(tenantID int) (<-chan []byte, func()) {
	ch := make(chan []byte, 16)

	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.subs[tenantID]
	if !ok {
		set = make(map[chan []byte]struct{})
		h.subs[tenantID] = set
	}

	set[ch] = struct{}{}

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if set, ok := h.subs[tenantID]; ok {
			delete(set, ch)
		}

		close(ch)
	}
}

// Publish delivers msg to all subscribers of the tenant.
// A full buffer drops the message: streams stay live, slow readers
// miss events instead of stalling everyone.
func (h *Hub) Publish(tenantID int, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs := h.subs[tenantID]

	slog.Debug("publishing event", "tenant", tenantID, "subscribers", len(subs))

	for ch := range subs {
		select {
		case ch <- msg:
		default:
			slog.Debug("dropping event for slow subscriber", "tenant", tenantID)
		}
	}
}
