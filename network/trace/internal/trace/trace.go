// Package trace carries W3C trace-context across service hops:
// parse incoming traceparent, or start a new trace when absent.
package trace

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const Header = "traceparent"

// Context is one hop of a distributed trace: the shared trace ID
// plus this hop's parent (span) ID.
type Context struct {
	TraceID  string
	ParentID string
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

// New starts a fresh trace (version 00, sampled flag 01).
func New() Context {
	return Context{TraceID: randomHex(16), ParentID: randomHex(8)}
}

// Extract reads traceparent, falling back to a new trace when the
// header is missing or malformed. Never fails the request.
func Extract(r *http.Request) Context {
	parts := strings.Split(r.Header.Get(Header), "-")
	if len(parts) != 4 || len(parts[1]) != 32 || len(parts[2]) != 16 {
		return New()
	}

	return Context{TraceID: parts[1], ParentID: parts[2]}
}

// Child keeps the trace ID, mints a new span ID for the next hop.
func (c Context) Child() Context {
	return Context{TraceID: c.TraceID, ParentID: randomHex(8)}
}

// Value renders the W3C traceparent value for the next hop.
func (c Context) Value() string {
	return "00-" + c.TraceID + "-" + c.ParentID + "-01"
}
