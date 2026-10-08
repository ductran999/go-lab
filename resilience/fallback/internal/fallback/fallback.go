// Package fallback runs the degrade ladder per call: fresh truth,
// stale last-good inside its age bound, static default, honest
// error. Strict mode skips stale and default (core paths: wrong
// beats error never). Bounds are the whole technique — unbounded
// fallback is a lie with a counter.
package fallback

import (
	"errors"
	"sync"
	"time"
)

// ErrDownstream marks primary failure after the ladder gave out.
var ErrDownstream = errors.New("downstream failed")

// Sources served, in ladder order.
const (
	SourceFresh      = "fresh"
	SourceStale      = "stale"
	SourceStaleRedis = "stale-redis"
	SourceDefault    = "default"
)

// storeKey is the single shared copy (demo: one resource).
const storeKey = "fallback:last-good"

// Store is a shared last-good copy (Redis): survives restarts,
// visible across instances. Errors mean "no copy", never fatal —
// the ladder degrades past a dead store without a sound.
type Store interface {
	Get(key string) (string, bool)
	Set(key, val string, ttl time.Duration)
}

// Ladder holds the last-good copy and the bounds. Zero value is
// strict with no default (fail fast); set fields for the edge.
// Shared adds the L2 Redis step between L1 memory and default.
type Ladder struct {
	Bound   time.Duration
	Strict  bool
	Default string
	Shared  Store

	mu   sync.Mutex
	body string
	at   time.Time
}

// Call runs fetch; on failure it degrades down the ladder and
// reports which step served (for X-Source headers and counters).
func (l *Ladder) Call(fetch func() (string, error)) (string, string, error) {
	body, err := fetch()
	if err == nil {
		l.save(body)

		return body, SourceFresh, nil
	}

	if l.Strict {
		return "", "", err
	}

	if body, ok := l.stale(); ok {
		return body, SourceStale, nil
	}

	if body, ok := l.shared(); ok {
		return body, SourceStaleRedis, nil
	}

	if l.Default != "" {
		return l.Default, SourceDefault, nil
	}

	return "", "", err
}

// save stores the fresh copy in L1 and shares it with L2.
func (l *Ladder) save(body string) {
	l.mu.Lock()
	l.body, l.at = body, time.Now()
	l.mu.Unlock()

	if l.Shared != nil {
		l.Shared.Set(storeKey, body, l.Bound)
	}
}

// stale returns the L1 copy inside its age bound.
func (l *Ladder) stale() (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.body == "" || time.Since(l.at) >= l.Bound {
		return "", false
	}

	return l.body, true
}

// shared returns the L2 copy (miss, expiry, and dead store all
// read as absent — the ladder steps past them silently).
func (l *Ladder) shared() (string, bool) {
	if l.Shared == nil {
		return "", false
	}

	return l.Shared.Get(storeKey)
}
