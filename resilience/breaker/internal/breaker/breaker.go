// Package breaker is a stdlib circuit breaker: closed counts
// consecutive retryable failures, open fails fast for cooldown,
// half-open admits exactly one probe. Production maps 1:1 to
// sony/gobreaker (Counts, ReadyToTrip, OnStateChange) — same
// states, same rules, borrowed wheel when the lab grows up.
package breaker

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// ErrOpen marks fail-fast rejection: open circuit, or another
// flight already probing. Callers treat it as "don't retry into
// this" — backing off is the breaker's job, not theirs.
var ErrOpen = errors.New("circuit open")

// StatusError carries the HTTP status of a failed call so the
// breaker counts retryable failures only.
type StatusError struct {
	Status int
	URL    string
}

// Error implements error.
func (e *StatusError) Error() string {
	return fmt.Sprintf("get %s: status %d", e.URL, e.Status)
}

// RetryableStatus counts timeouts, resets (plain errors), 429,
// 408, and 5xx — never 4xx verdicts or cancelled contexts.
func RetryableStatus(err error) bool {
	var status *StatusError
	if !errors.As(err, &status) {
		return true
	}

	if status.Status == http.StatusTooManyRequests ||
		status.Status == http.StatusRequestTimeout ||
		status.Status >= http.StatusInternalServerError {
		return true
	}

	return false
}

// State is the breaker position.
type State int

const (
	// StateClosed lets traffic through and counts failures.
	StateClosed State = iota
	// StateOpen rejects everything until cooldown passes.
	StateOpen
	// StateHalfOpen admits one probe flight.
	StateHalfOpen
)

// String names the state for logs and demos.
func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Breaker guards one downstream. Threshold counts consecutive
// retryable failures; Cooldown parks the circuit open; IsFailure
// classifies (nil counts every error — pass a predicate like
// retryable-status so 403s never trip the wire).
type Breaker struct {
	Threshold int
	Cooldown  time.Duration
	IsFailure func(error) bool

	mu      sync.Mutex
	state   State
	consec  int
	opened  time.Time
	probing bool

	calls    atomic.Int64
	rejected atomic.Int64
	probes   atomic.Int64
}

// New builds a closed breaker (threshold floors at 1).
func New(threshold int, cooldown time.Duration, isFailure func(error) bool) *Breaker {
	if threshold < 1 {
		threshold = 1
	}

	return &Breaker{Threshold: threshold, Cooldown: cooldown, IsFailure: isFailure}
}

// State reports the current position (for demos and dashboards).
func (b *Breaker) State() State {
	b.mu.Lock()

	defer b.mu.Unlock()

	return b.state
}

// Stats reports executed calls, fast rejections, and probe flights.
func (b *Breaker) Stats() (int64, int64, int64) {
	return b.calls.Load(), b.rejected.Load(), b.probes.Load()
}

// Call runs fn unless the circuit refuses. Refusals return ErrOpen
// (never fn's error — nothing executed). Half-open admits one probe;
// a second concurrent call is refused even though the state reads
// half-open — one probe means one.
func (b *Breaker) Call(fn func() error) error {
	b.mu.Lock()

	probe := false

	switch b.state {
	case StateOpen:
		if time.Since(b.opened) < b.Cooldown {
			b.mu.Unlock()
			b.rejected.Add(1)

			return ErrOpen
		}

		b.state = StateHalfOpen
		b.probing = true
		probe = true
	case StateHalfOpen:
		if b.probing {
			b.mu.Unlock()
			b.rejected.Add(1)

			return ErrOpen
		}

		b.probing = true
		probe = true
	case StateClosed:
		// healthy path: straight to execution below
	}

	b.mu.Unlock()
	b.calls.Add(1)

	if probe {
		b.probes.Add(1)
	}

	err := fn()
	failed := err != nil && (b.IsFailure == nil || b.IsFailure(err))

	b.mu.Lock()

	defer b.mu.Unlock()

	if b.state == StateHalfOpen {
		b.probing = false

		if failed {
			b.state = StateOpen
			b.opened = time.Now()
			b.consec = 0
		} else {
			b.state = StateClosed
			b.consec = 0
		}

		return err
	}

	if b.state != StateClosed {
		return err
	}

	if failed {
		b.consec++

		if b.consec >= b.Threshold {
			b.state = StateOpen
			b.opened = time.Now()
			b.consec = 0
		}
	} else {
		b.consec = 0
	}

	return err
}
