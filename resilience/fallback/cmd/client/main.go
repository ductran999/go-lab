// Command client walks the ladder against a killable primary:
// fresh while it answers, stale last-good inside the age bound,
// static default past it, honest error in strict mode. Each line
// prints its source — the dashboard equivalent of X-Source.
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

	"go-lab/resilience/fallback/internal/fallback"
)

// errBadStatus marks a non-200 primary answer.
var errBadStatus = errors.New("bad status")

func main() {
	base := flag.String("base", "http://localhost:8127", "fallback lab server")
	n := flag.Int("n", 3, "calls to make")
	stale := flag.Duration("stale", 2*time.Minute, "stale age bound")
	strict := flag.Bool("strict", false, "skip stale and default (core paths)")
	def := flag.String("default", "unknown", "static default value")
	redisAddr := flag.String("redis", "localhost:6379", "shared copy addr (empty: L1 only)")
	gap := flag.Duration("gap", 0, "sleep between calls (to age the copy past -stale)")

	flag.Parse()

	err := run(*base, *n, *stale, *strict, *def, *redisAddr, *gap)
	if err != nil {
		fmt.Println("client:", err)
		os.Exit(1)
	}
}

func run(base string, n int, bound time.Duration, strict bool, def, redisAddr string, gap time.Duration) error {
	var shared fallback.Store
	if redisAddr != "" {
		shared = fallback.NewRedisStore(redisAddr)
	}

	l := &fallback.Ladder{Bound: bound, Strict: strict, Default: def, Shared: shared}

	counts := map[string]int{}

	for range n {
		body, src, err := l.Call(func() (string, error) {
			return get(context.Background(), base+"/flaky")
		})
		if err != nil {
			fmt.Printf("error: %v\n", err)

			counts["error"]++

			continue
		}

		fmt.Printf("%s: %s\n", src, body)
		counts[src]++

		time.Sleep(gap)
	}

	fmt.Printf("sources: %v\n", counts)

	return nil
}

// get GETs url with a short deadline (fallback triggers on slow too).
func get(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)

	defer cancel()

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
		return "", fmt.Errorf("%w: get %s: status %d", errBadStatus, url, resp.StatusCode)
	}

	return string(raw), nil
}
