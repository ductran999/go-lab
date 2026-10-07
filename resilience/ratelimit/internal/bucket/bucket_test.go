package bucket

import (
	"context"
	"testing"
	"time"
)

func TestBucketBurstsThenRefuses(t *testing.T) {
	t.Parallel()

	b := NewBucket(10, 2)

	if !b.Take() {
		t.Fatal("first token of burst must admit")
	}

	if !b.Take() {
		t.Fatal("second token of burst must admit")
	}

	if b.Take() {
		t.Fatal("empty bucket must refuse, not queue")
	}

	time.Sleep(250 * time.Millisecond)

	if !b.Take() {
		t.Fatal("250ms at 10/s must refill 2.5 tokens")
	}
}

func TestWaitPacesAndRespectsCancel(t *testing.T) {
	t.Parallel()

	b := NewBucket(10, 1)

	if !b.Wait(context.Background()) {
		t.Fatal("full tank must admit at once")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if b.Wait(ctx) {
		t.Fatal("cancelled ctx must shed, not wait for refill")
	}
}
