package fallback

import (
	"errors"
	"testing"
	"time"
)

var errDead = errors.New("dead")

func TestFreshStoresLastGood(t *testing.T) {
	t.Parallel()

	l := &Ladder{Bound: time.Minute, Default: "dflt"}

	body, src, err := l.Call(func() (string, error) { return "v1", nil })
	if err != nil || src != SourceFresh || body != "v1" {
		t.Fatalf("got %q %q %v", body, src, err)
	}
}

func TestStaleWithinBound(t *testing.T) {
	t.Parallel()

	l := &Ladder{Bound: time.Minute, Default: "dflt"}

	_, _, err := l.Call(func() (string, error) { return "v1", nil })
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	body, src, err := l.Call(func() (string, error) { return "", errDead })
	if err != nil || src != SourceStale || body != "v1" {
		t.Fatalf("got %q %q %v, want stale v1", body, src, err)
	}
}

func TestBoundExpiryFallsToDefault(t *testing.T) {
	t.Parallel()

	l := &Ladder{Bound: 20 * time.Millisecond, Default: "dflt"}

	_, _, err := l.Call(func() (string, error) { return "v1", nil })
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	body, src, err := l.Call(func() (string, error) { return "", errDead })
	if err != nil || src != SourceDefault || body != "dflt" {
		t.Fatalf("got %q %q %v, want default", body, src, err)
	}
}

func TestStrictFailsFast(t *testing.T) {
	t.Parallel()

	l := &Ladder{Strict: true, Bound: time.Minute, Default: "dflt"}

	_, _, err := l.Call(func() (string, error) { return "v1", nil })
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, _, err = l.Call(func() (string, error) { return "", errDead })
	if !errors.Is(err, errDead) {
		t.Fatalf("err = %v, strict must surface the real failure", err)
	}
}

func TestEmptyLadderErrors(t *testing.T) {
	t.Parallel()

	l := &Ladder{}

	_, _, err := l.Call(func() (string, error) { return "", errDead })
	if !errors.Is(err, errDead) {
		t.Fatalf("err = %v, want the original failure", err)
	}
}

// fakeStore is a scripted Store: values plus failure injection.
type fakeStore struct {
	vals map[string]string
	gets int
	fail bool
}

func (f *fakeStore) Get(key string) (string, bool) {
	f.gets++

	if f.fail {
		return "", false
	}

	v, ok := f.vals[key]

	return v, ok
}

func (f *fakeStore) Set(key, val string, _ time.Duration) {
	if f.vals == nil {
		f.vals = map[string]string{}
	}

	f.vals[key] = val
}

func TestSharedSavesOnFresh(t *testing.T) {
	t.Parallel()

	store := &fakeStore{}
	l := &Ladder{Bound: time.Minute, Shared: store}

	_, _, err := l.Call(func() (string, error) { return "v1", nil })
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	if store.vals[storeKey] != "v1" {
		t.Fatalf("shared copy = %q, want v1", store.vals[storeKey])
	}
}

func TestSharedServesColdProcess(t *testing.T) {
	t.Parallel()

	store := &fakeStore{vals: map[string]string{storeKey: "v9"}}

	cold := &Ladder{Bound: time.Minute, Default: "dflt", Shared: store}

	body, src, err := cold.Call(func() (string, error) { return "", errDead })
	if err != nil || src != SourceStaleRedis || body != "v9" {
		t.Fatalf("got %q %q %v, want stale-redis v9", body, src, err)
	}
}

func TestSharedFailureFallsToDefault(t *testing.T) {
	t.Parallel()

	store := &fakeStore{fail: true}
	l := &Ladder{Bound: time.Minute, Default: "dflt", Shared: store}

	body, src, err := l.Call(func() (string, error) { return "", errDead })
	if err != nil || src != SourceDefault || body != "dflt" {
		t.Fatalf("got %q %q %v, want default past a dead store", body, src, err)
	}
}

func TestStrictSkipsShared(t *testing.T) {
	t.Parallel()

	store := &fakeStore{vals: map[string]string{storeKey: "v9"}}
	l := &Ladder{Strict: true, Bound: time.Minute, Default: "dflt", Shared: store}

	_, _, err := l.Call(func() (string, error) { return "", errDead })
	if !errors.Is(err, errDead) {
		t.Fatalf("err = %v, strict must surface the failure", err)
	}

	if store.gets != 0 {
		t.Fatalf("shared gets = %d, strict must not touch L2", store.gets)
	}
}
