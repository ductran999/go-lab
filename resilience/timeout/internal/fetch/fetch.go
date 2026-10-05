// Package fetch is a timeout-first HTTP client: every call takes a
// context, slow upstreams die by deadline, and the tail gets hedged
// (a second request fires when the first looks sick; first success
// wins, the loser is cancelled).
package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrUpstream marks any failure beyond our control: DNS, reset,
// timeout, or a non-200 status. Errors wrap it, so errors.Is works.
var ErrUpstream = errors.New("upstream failed")

// Do GETs url under ctx. No context, no call: the deadline belongs
// to the caller, and it propagates to the server (which stops work
// on ctx.Done instead of burning CPU for a client that's gone).
func Do(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request %s: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: get %s: %w", ErrUpstream, url, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: get %s: status %d", ErrUpstream, url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: read %s: %w", ErrUpstream, url, err)
	}

	return string(body), nil
}

// DoHedged GETs primary, then fires hedge if primary hasn't answered
// within hedgeAfter. First success wins; a fast failure falls through
// to the other flight; both failing joins the errors. Either way the
// loser dies by context cancel — hedging without cancel is just 2x load.
//
// Hedge idempotent reads only: a second POST may double-charge.
func DoHedged(ctx context.Context, primary, hedge string, hedgeAfter time.Duration) (string, error) {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	type result struct {
		body string
		err  error
	}

	out := make(chan result, 2)

	go func() {
		body, err := Do(ctx, primary)
		out <- result{body: body, err: err}
	}()

	timer := time.NewTimer(hedgeAfter)

	defer timer.Stop()

	select {
	case res := <-out:
		return res.body, res.err
	case <-timer.C:
	}

	go func() {
		body, err := Do(ctx, hedge)
		out <- result{body: body, err: err}
	}()

	first := <-out
	if first.err == nil {
		return first.body, nil
	}

	second := <-out
	if second.err != nil {
		return "", errors.Join(first.err, second.err)
	}

	return second.body, nil
}
