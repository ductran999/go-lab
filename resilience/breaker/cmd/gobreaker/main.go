// Command gobreaker runs the same dead-then-healed lifecycle as
// client, but on sony/gobreaker instead of the stdlib breaker:
// same states, same rules, borrowed wheel. OnStateChange prints
// every transition — including half-open, which the stdlib demo
// only implies. Compare outputs of the two clients: identical story.
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

	"github.com/sony/gobreaker"

	"go-lab/resilience/breaker/internal/breaker"
)

func main() {
	base := flag.String("base", "http://localhost:8124", "breaker lab server")
	threshold := flag.Uint64("threshold", 5, "consecutive failures to open")
	cooldown := flag.Duration("cooldown", 5*time.Second, "open duration before probe")

	flag.Parse()

	err := run(*base, *threshold, *cooldown)
	if err != nil {
		fmt.Println("gobreaker client:", err)
		os.Exit(1)
	}
}

func run(base string, threshold uint64, cooldown time.Duration) error {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "flaky",
		MaxRequests: 1,
		Timeout:     cooldown,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return uint64(counts.ConsecutiveFailures) >= threshold
		},
		OnStateChange: func(_ string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("state: %s -> %s\n", from, to)
		},
	})

	err := reset(base, 100)
	if err != nil {
		return err
	}

	fmt.Println("--- phase 1: downstream dead, hammering")

	hammer(base, cb, 8*time.Second, 200*time.Millisecond)

	err = reset(base, 0)
	if err != nil {
		return err
	}

	fmt.Println("--- phase 2: downstream healed, waiting for the probe")

	hammer(base, cb, cooldown+4*time.Second, 500*time.Millisecond)

	fmt.Printf("final: state=%s counts=%+v\n", cb.State(), cb.Counts())

	return nil
}

// hammer calls /flaky through the breaker, printing fast rejections
// (gobreaker.ErrOpen) the same way the stdlib demo does.
func hammer(base string, cb *gobreaker.CircuitBreaker, dur, every time.Duration) {
	deadline := time.Now().Add(dur)

	for time.Now().Before(deadline) {
		_, err := cb.Execute(func() (any, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			defer cancel()

			return getBody(ctx, base+"/flaky")
		})
		if errors.Is(err, gobreaker.ErrOpenState) {
			fmt.Printf("rejected (state %s)\n", cb.State())
		}

		time.Sleep(every)
	}
}

// reset rearms the server failure budget.
func reset(base string, fail int) error {
	_, err := getBody(context.Background(), fmt.Sprintf("%s/reset?fail=%d", base, fail))

	return err
}

// getBody GETs url. Non-200 answers come home as breaker.StatusError
// (this demo server emits 500s only, so both breakers behave the
// same; note gobreaker v1 counts every error while the stdlib one
// skips non-retryable via RetryableStatus — see README).
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
