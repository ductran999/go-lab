// Command client proves lane isolation: slow and fast traffic fire
// at once — with separate lanes fast stays fast while slow saturates
// its own pool; one shared lane drags fast behind slow.
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

	"go-lab/resilience/bulkhead/internal/guard"
)

func main() {
	base := flag.String("base", "http://localhost:8125", "bulkhead lab server")
	split := flag.Bool("split", true, "separate lanes per downstream (false: one shared lane)")

	flag.Parse()

	err := run(*base, *split)
	if err != nil {
		fmt.Println("client:", err)
		os.Exit(1)
	}
}

func run(base string, split bool) error {
	return bulkhead(base, split)
}

// bulkhead fires 6 slow (1s) + 6 fast calls at once and reports the
// slowest fast call: split lanes keep it at milliseconds, one shared
// lane of 4 parks fast behind slow for a full second.
func bulkhead(base string, split bool) error {
	slow, fast := guard.NewPool(2), guard.NewPool(2)
	shared := guard.NewPool(4)

	var worstFast, fastErr, fastOk, slowOk atomic.Int64

	var wg sync.WaitGroup

	start := time.Now()

	for range 6 {
		wg.Go(func() {
			lane := laneFor(shared, slow, split)
			lane.Acquire()

			defer lane.Release()

			get(context.Background(), base+"/slow?ms=1000")
			slowOk.Add(1)
		})
	}

	// Let slow traffic saturate the lanes before fast arrives: the
	// shared lane is then provably full, the fast lane provably free.
	time.Sleep(300 * time.Millisecond)

	for range 6 {
		wg.Go(func() {
			t := time.Now()

			lane := laneFor(shared, fast, split)
			lane.Acquire()

			defer lane.Release()

			code, _ := getCode(context.Background(), base+"/fast")

			ms := time.Since(t).Milliseconds()

			if code != http.StatusOK {
				fastErr.Add(1)
			} else {
				fastOk.Add(1)
			}

			for {
				old := worstFast.Load()
				if ms <= old || worstFast.CompareAndSwap(old, ms) {
					break
				}
			}
		})
	}

	wg.Wait()

	fmt.Printf("bulkhead split=%v lanes=%s slow=%d fast=%d/%d slowest-fast=%dms wall=%v\n",
		split, lanes(split), slowOk.Load(), fastOk.Load(), fastErr.Load(),
		worstFast.Load(), time.Since(start).Round(time.Millisecond))

	return nil
}

// lanes describes the partition for the log line.
func lanes(split bool) string {
	if split {
		return "slow:2+fast:2"
	}

	return "shared:4"
}

// laneFor picks the shared lane or the dedicated one.
func laneFor(shared, dedicated *guard.Pool, split bool) *guard.Pool {
	if split {
		return dedicated
	}

	return shared
}

// getCode GETs url and returns its status (0 on transport failure).
func getCode(ctx context.Context, url string) (int, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, ""
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.StatusCode, resp.Header.Get("Retry-After")
}

// ratelimit is a separate lab (resilience/ratelimit): the bucket
// lives client-side there, throttling my own calls.

// get GETs url, discarding the body (demo cares about timing, not content).
func get(ctx context.Context, url string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	_, _ = io.Copy(io.Discard, resp.Body)
}
