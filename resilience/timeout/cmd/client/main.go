// Command client runs three calls against the timeout lab server
// and prints what each cost: no deadline (pays full 2s), a tight
// deadline (fails fast), and hedging (a second flight after 300ms
// cuts the tail without paying the full wait).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"go-lab/resilience/timeout/internal/fetch"
)

func main() {
	base := flag.String("base", "http://localhost:8121", "timeout lab server")

	flag.Parse()

	slow := *base + "/slow?ms=2000"
	fast := *base + "/fast"

	timed("1. no deadline", func() error {
		_, err := fetch.Do(context.Background(), slow)

		return err
	})

	timed("2. 300ms deadline", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)

		defer cancel()

		_, err := fetch.Do(ctx, slow)

		return err
	})

	timed("3. hedged after 300ms", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		body, err := fetch.DoHedged(ctx, slow, fast, 300*time.Millisecond)
		if err == nil {
			fmt.Printf("   body: %s\n", body)
		}

		return err
	})
}

// timed runs fn and prints its wall time plus the short error.
func timed(name string, fn func() error) {
	start := time.Now()

	err := fn()

	fmt.Printf("%s: took %v", name, time.Since(start).Round(time.Millisecond))

	if err != nil {
		short := err
		if errors.Is(err, context.DeadlineExceeded) {
			short = context.DeadlineExceeded
		}

		fmt.Printf(" err=%v", short)
	}

	fmt.Println()

	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		os.Exit(1)
	}
}
