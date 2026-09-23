package onmem

import (
	"github.com/dgraph-io/ristretto/v2"
)

func NewRistretto() (*ristretto.Cache[string, string], error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	return cache, nil
}
