package guard

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolCapsConcurrency(t *testing.T) {
	t.Parallel()

	p := NewPool(2)

	var inside, peak atomic.Int64

	var wg sync.WaitGroup

	for range 10 {
		wg.Go(func() {
			p.Acquire()

			defer p.Release()

			n := inside.Add(1)

			for {
				old := peak.Load()
				if n <= old || peak.CompareAndSwap(old, n) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			inside.Add(-1)
		})
	}

	wg.Wait()

	if got := peak.Load(); got != 2 {
		t.Fatalf("peak concurrency = %d, want exactly the pool size", got)
	}
}

func TestSlowPoolNeverBlocksFastLane(t *testing.T) {
	t.Parallel()

	slow, fast := NewPool(1), NewPool(1)

	slow.Acquire()

	defer slow.Release()

	done := make(chan struct{})

	go func() {
		defer close(done)

		fast.Acquire()

		defer fast.Release()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("fast lane waited behind the saturated slow pool")
	}
}
