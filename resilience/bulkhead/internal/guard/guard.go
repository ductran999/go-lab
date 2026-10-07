// Package guard holds the bulkhead primitive: Pool partitions
// concurrency per downstream (own lane, bounded). Slow downstream
// saturates its own pool; other pools never wait.
package guard

// Pool is a counting semaphore: at most n holders inside.
// Slow downstream saturates its own pool; other pools never wait.
type Pool struct {
	sem chan struct{}
}

// NewPool builds a lane of n slots.
func NewPool(n int) *Pool {
	if n < 1 {
		n = 1
	}

	return &Pool{sem: make(chan struct{}, n)}
}

// Acquire takes one slot (blocks until free).
func (p *Pool) Acquire() {
	p.sem <- struct{}{}
}

// Release frees one slot.
func (p *Pool) Release() {
	<-p.sem
}
