// Command client stampedes /hot with n concurrent identical GETs.
// With -fly the x/sync/singleflight group collapses them into one
// origin query (shared=N-1); without it every caller pays full
// price (queries=N). Same key, same read — the only safe shape.
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

	"golang.org/x/sync/singleflight"
)

// errBadFlight marks a flight that returned the wrong type.
var errBadFlight = errors.New("bad flight type")

// errBadStatus marks a non-200 origin answer.
var errBadStatus = errors.New("bad status")

func main() {
	base := flag.String("base", "http://localhost:8123", "singleflight lab server")
	n := flag.Int("n", 20, "concurrent identical callers")
	fly := flag.Bool("fly", true, "collapse with singleflight")

	flag.Parse()

	err := run(*base, *n, *fly)
	if err != nil {
		fmt.Println("client:", err)
		os.Exit(1)
	}
}

func run(base string, n int, fly bool) error {
	var group singleflight.Group

	var ok, failed, shared atomic.Int64

	var wg sync.WaitGroup

	start := time.Now()

	for range n {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			defer cancel()

			body, wasShared, err := fetch(ctx, &group, fly, base+"/hot?ms=300")
			if err != nil {
				failed.Add(1)

				return
			}

			if wasShared {
				shared.Add(1)
			}

			if body == "" {
				failed.Add(1)

				return
			}

			ok.Add(1)
		})
	}

	wg.Wait()

	stats, err := getBody(context.Background(), base+"/stats")
	if err != nil {
		return err
	}

	fmt.Printf("clients=%d fly=%v ok=%d failed=%d shared=%d wall=%v server=%s\n",
		n, fly, ok.Load(), failed.Load(), shared.Load(), time.Since(start).Round(time.Millisecond), stats)

	return nil
}

// fetch GETs url directly, or shares one flight per key under fly.
func fetch(ctx context.Context, group *singleflight.Group, fly bool, url string) (string, bool, error) {
	if !fly {
		body, err := getBody(ctx, url)

		return body, false, err
	}

	v, err, shared := group.Do("hot-item", func() (any, error) {
		return getBody(ctx, url)
	})
	if err != nil {
		return "", shared, err
	}

	body, ok := v.(string)
	if !ok {
		return "", shared, fmt.Errorf("%w: flight returned %T", errBadFlight, v)
	}

	return body, shared, nil
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: get %s: status %d", errBadStatus, url, resp.StatusCode)
	}

	return string(raw), nil
}
