// Command race fires three flights at once — /slow?ms=2000,
// /slow?ms=800, /fast — and prints the winner. The two losers die
// by cancel: /stats shows them as abandoned, which is the price of
// the shortest tail physically possible.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go-lab/resilience/timeout/internal/fetch"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("race:", err)
		os.Exit(1)
	}
}

// run races three flights and prints the winner.
func run() error {
	base := flag.String("base", "http://localhost:8121", "timeout lab server")

	flag.Parse()

	urls := []string{*base + "/slow?ms=2000", *base + "/slow?ms=800", *base + "/fast"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	start := time.Now()

	body, winner, err := fetch.DoFastest(ctx, urls...)
	if err != nil {
		return err
	}

	fmt.Printf("winner: flight %d (%s) in %v: %s\n", winner, urls[winner], time.Since(start).Round(time.Millisecond), body)

	return nil
}
