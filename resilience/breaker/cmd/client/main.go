// Command client runs the breaker lifecycle hands-free: kill the
// downstream (fail=100), hammer until the circuit opens (fast
// rejects, zero downstream load), heal it (fail=0), and watch the
// half-open probe close the circuit by itself. State transitions
// and the final counters are the whole demo.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go-lab/resilience/breaker/internal/breaker"
)

func main() {
	base := flag.String("base", "http://localhost:8124", "breaker lab server")
	threshold := flag.Int("threshold", 5, "consecutive failures to open")
	cooldown := flag.Duration("cooldown", 5*time.Second, "open duration before probe")

	flag.Parse()

	err := run(*base, *threshold, *cooldown)
	if err != nil {
		fmt.Println("client:", err)
		os.Exit(1)
	}
}

func run(base string, threshold int, cooldown time.Duration) error {
	b := breaker.New(threshold, cooldown, breaker.RetryableStatus)

	err := reset(base, 100)
	if err != nil {
		return err
	}

	fmt.Println("--- phase 1: downstream dead, hammering")

	hammer(base, b, 8*time.Second, 200*time.Millisecond)

	err = reset(base, 0)
	if err != nil {
		return err
	}

	fmt.Println("--- phase 2: downstream healed, waiting for the probe")

	hammer(base, b, cooldown+4*time.Second, 500*time.Millisecond)

	calls, rejected, probes := b.Stats()

	fmt.Printf("final: state=%s calls=%d rejected=%d probes=%d\n", b.State(), calls, rejected, probes)

	return nil
}

// hammer calls /flaky through the breaker until dur passes, printing
// only state transitions and refusals (success lines would flood).
func hammer(base string, b *breaker.Breaker, dur, every time.Duration) {
	last := breaker.StateClosed

	deadline := time.Now().Add(dur)

	for time.Now().Before(deadline) {
		err := b.Call(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			defer cancel()

			_, err := getBody(ctx, base+"/flaky")

			return err
		})

		state := b.State()
		if state != last {
			fmt.Printf("state: %s -> %s\n", last, state)
			last = state
		}

		if errors.Is(err, breaker.ErrOpen) {
			fmt.Printf("rejected (state %s)\n", state)
		}

		time.Sleep(every)
	}
}

// reset rearms the server failure budget.
func reset(base string, fail int) error {
	_, err := getBody(context.Background(), fmt.Sprintf("%s/reset?fail=%d", base, fail))

	return err
}

// getBody GETs url and returns it as string. Non-200 answers come
// home as breaker.StatusError so the breaker counts retryable
// failures only (a 403 never trips the wire).
func getBody(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get %s: %w", url, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &breaker.StatusError{Status: resp.StatusCode, URL: url}
	}

	return string(raw), nil
}
