package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/singleflight"
)

func TestCollapseSharesOneFlight(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		time.Sleep(100 * time.Millisecond)

		_, _ = w.Write([]byte("same-body"))
	}))

	defer srv.Close()

	var group singleflight.Group

	const n = 20

	bodies := make([]string, n)
	shared := make([]bool, n)

	var wg sync.WaitGroup

	for i := range n {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			defer cancel()

			body, wasShared, err := fetch(ctx, &group, true, srv.URL)
			if err != nil {
				t.Errorf("caller: %v", err)

				return
			}

			bodies[i] = body
			shared[i] = wasShared
		})
	}

	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("origin queries = %d, want 1 shared flight", got)
	}

	followers := 0

	for i := range n {
		if bodies[i] != "same-body" {
			t.Fatalf("caller %d body = %q, want shared answer", i, bodies[i])
		}

		if shared[i] {
			followers++
		}
	}

	if followers != n {
		t.Fatalf("shared = %d, want %d (leader reports shared too when followers attach)", followers, n)
	}
}

func TestDirectPaysEveryCaller(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)

		_, _ = w.Write([]byte("ok"))
	}))

	defer srv.Close()

	var group singleflight.Group

	const n = 10

	var wg sync.WaitGroup

	for range n {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			defer cancel()

			_, _, err := fetch(ctx, &group, false, srv.URL)
			if err != nil {
				t.Errorf("caller: %v", err)
			}
		})
	}

	wg.Wait()

	if got := calls.Load(); got != n {
		t.Fatalf("origin queries = %d, want %d (no sharing)", got, n)
	}
}

func TestSharedErrorReachesFollowers(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer srv.Close()

	var group singleflight.Group

	const n = 5

	var wg sync.WaitGroup

	errs := make([]error, n)

	for i := range n {
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			defer cancel()

			_, _, err := fetch(ctx, &group, true, srv.URL)
			errs[i] = err
		})
	}

	wg.Wait()

	for i := range n {
		if !errors.Is(errs[i], errBadStatus) {
			t.Fatalf("follower %d err = %v, want the shared 500", i, errs[i])
		}
	}
}
