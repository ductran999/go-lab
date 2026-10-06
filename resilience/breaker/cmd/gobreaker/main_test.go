package main

import (
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
)

var errBoom = errors.New("boom")

// TestGobreakerParity proves the borrowed wheel tells the same story:
// opens after threshold, rejects while open, half-open probe closes.
func TestGobreakerParity(t *testing.T) {
	t.Parallel()

	var states []gobreaker.State

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "parity",
		MaxRequests: 1,
		Timeout:     30 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
		OnStateChange: func(_ string, _ gobreaker.State, to gobreaker.State) {
			states = append(states, to)
		},
	})

	fail := func() (any, error) { return nil, errBoom }

	_, err := cb.Execute(fail)
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	_, err = cb.Execute(fail)
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	_, err = cb.Execute(fail)
	if !errors.Is(err, gobreaker.ErrOpenState) {
		t.Fatalf("err = %v, want open rejection", err)
	}

	time.Sleep(40 * time.Millisecond)

	v, err := cb.Execute(func() (any, error) { return "healed", nil })
	if err != nil || v.(string) != "healed" {
		t.Fatalf("probe = %v, %v", v, err)
	}

	want := []gobreaker.State{gobreaker.StateOpen, gobreaker.StateHalfOpen, gobreaker.StateClosed}
	if len(states) != len(want) {
		t.Fatalf("transitions = %v, want %v", states, want)
	}

	for i := range want {
		if states[i] != want[i] {
			t.Fatalf("transitions = %v, want %v", states, want)
		}
	}
}
