// Command client hammers /flaky with n concurrent retrying clients,
// then prints the wall time and the server's per-50ms hit buckets.
// Without jitter the retries land in lockstep (tall narrow spikes);
// with jitter they spread (low wide humps). Same attempts, kinder load.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go-lab/resilience/retry/internal/retry"
)

func main() {
	base := flag.String("base", "http://localhost:8122", "retry lab server")
	n := flag.Int("n", 10, "concurrent clients")
	jitter := flag.Bool("jitter", true, "full jitter on backoff")
	attempts := flag.Int("attempts", 5, "tries per client")
	baseMs := flag.Int("baseMs", 50, "backoff base in ms")
	fail := flag.Int("fail", 10, "server fails the first N hits")
	path := flag.String("path", "flaky", "server endpoint (flaky, deny)")

	flag.Parse()

	err := reset(*base, *fail)
	if err != nil {
		fmt.Println("reset:", err)
		os.Exit(1)
	}

	var ok, failed atomic.Int64

	var wg sync.WaitGroup

	start := time.Now()

	for range *n {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			defer cancel()

			err := retry.Call(
				ctx,
				*attempts,
				time.Duration(*baseMs)*time.Millisecond,
				*jitter,
				retry.RetryableStatus,
				func(ctx context.Context) error {
					return get(ctx, *base+"/"+*path)
				},
			)
			if err != nil {
				failed.Add(1)

				return
			}

			ok.Add(1)
		})
	}

	wg.Wait()

	fmt.Printf("clients=%d jitter=%v ok=%d failed=%d wall=%v\n",
		*n, *jitter, ok.Load(), failed.Load(), time.Since(start).Round(time.Millisecond))

	stats, err := body(context.Background(), *base+"/stats")
	if err != nil {
		fmt.Println("stats:", err)
		os.Exit(1)
	}

	fmt.Println("server buckets (50ms each):", stats)
}

// reset rearms the server failure budget.
func reset(base string, fail int) error {
	return get(context.Background(), fmt.Sprintf("%s/reset?fail=%d", base, fail))
}

// get GETs url and errors on non-200 (a retryable blip by definition).
func get(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("drain %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return &retry.StatusError{Status: resp.StatusCode, URL: url}
	}

	return nil
}

// body GETs url and returns it as string (for /stats display).
func body(ctx context.Context, url string) (string, error) {
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

	return string(raw), nil
}
