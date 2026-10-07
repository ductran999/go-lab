// Command client paces itself through a token bucket: 30 calls at
// 10/sec (burst 10) take ~2s of wall instead of hitting the
// downstream all at once. The downstream sees smooth traffic;
// calls whose ctx dies waiting are shed, never queued past patience.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go-lab/resilience/ratelimit/internal/bucket"
)

// errBadStatus marks a non-200 downstream answer.
var errBadStatus = errors.New("bad status")

func main() {
	base := flag.String("base", "http://localhost:8126", "ratelimit lab server")
	n := flag.Int("n", 30, "concurrent calls")
	rate := flag.Float64("rate", 10, "bucket refill per second")
	burst := flag.Int("burst", 10, "bucket burst savings")

	flag.Parse()

	err := run(*base, *n, *rate, *burst)
	if err != nil {
		fmt.Println("client:", err)
		os.Exit(1)
	}
}

func run(base string, n int, rate float64, burst int) error {
	b := bucket.NewBucket(rate, burst)

	var ok, shed atomic.Int64

	var wg sync.WaitGroup

	start := time.Now()

	for range n {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			defer cancel()

			if !b.Wait(ctx) {
				shed.Add(1)

				return
			}

			err := get(ctx, base+"/work?ms=50")
			if err != nil {
				shed.Add(1)

				return
			}

			ok.Add(1)
		})
	}

	wg.Wait()

	fmt.Printf("ratelimit n=%d rate=%.0f burst=%d ok=%d shed=%d wall=%v (unthrottled wall would be ~ms)\n",
		n, rate, burst, ok.Load(), shed.Load(), time.Since(start).Round(time.Millisecond))

	stats, err := getBody(context.Background(), base+"/stats")
	if err != nil {
		return err
	}

	fmt.Println("server:", stats)

	return nil
}

// get GETs url (errors counted as shed by the caller).
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
		return fmt.Errorf("%w: get %s: status %d", errBadStatus, url, resp.StatusCode)
	}

	return nil
}

// getBody GETs url and returns it as string.
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

	return string(raw), nil
}
