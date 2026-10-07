// Package bucket caps my own call rate (token bucket): client-side
// throttle protecting the downstream. Take fails fast (shed);
// Wait blocks for the next token (pace). Same bucket, two policies.
package bucket

import (
	"context"
	"sync"
	"time"
)

// Bucket refills rate tokens/sec up to burst savings, full tank.
type Bucket struct {
	mu     sync.Mutex
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

// NewBucket builds rate/sec refilling up to burst savings.
func NewBucket(rate float64, burst int) *Bucket {
	if burst < 1 {
		burst = 1
	}

	return &Bucket{rate: rate, burst: burst, tokens: float64(burst), last: time.Now()}
}

// Take spends one token (true) or refuses (false). Never blocks.
func (b *Bucket) Take() bool {
	b.mu.Lock()

	defer b.mu.Unlock()

	now := time.Now()
	b.tokens += now.Sub(b.last).Seconds() * b.rate

	if b.tokens > float64(b.burst) {
		b.tokens = float64(b.burst)
	}

	b.last = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--

	return true
}

// RetryAfter reports how long until the next token exists.
func (b *Bucket) RetryAfter() time.Duration {
	b.mu.Lock()

	defer b.mu.Unlock()

	if b.tokens >= 1 {
		return 0
	}

	secs := (1 - b.tokens) / b.rate
	if secs < 0 {
		return 0
	}

	return time.Duration(secs * float64(time.Second))
}

// Wait blocks for the next token (pacing) or ctx death, whichever
// comes first. False means the caller gave up — shed, don't queue
// past the caller's patience.
func (b *Bucket) Wait(ctx context.Context) bool {
	for {
		if b.Take() {
			return true
		}

		wait := b.RetryAfter()
		if wait <= 0 {
			wait = time.Millisecond
		}

		timer := time.NewTimer(wait)

		select {
		case <-ctx.Done():
			timer.Stop()

			return false
		case <-timer.C:
		}
	}
}
