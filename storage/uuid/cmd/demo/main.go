// Command demo prints v4 vs v7 side by side: structure, sort order,
// and what each leaks. Usage: go run ./cmd/demo
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrBadTimestamp = errors.New("timestamp out of range")

func fail(err error) {
	fmt.Fprintln(os.Stderr, "demo:", err)

	os.Exit(1)
}

func ordered(ids []string) bool {
	for i := 1; i < len(ids); i++ {
		if ids[i-1] > ids[i] {
			return false
		}
	}

	return true
}

// v7BornAt decodes the creation time hidden in a v7 prefix:
// first 12 hex chars = unix milliseconds.
func v7BornAt(id string) (time.Time, error) {
	raw := strings.ReplaceAll(id[:13], "-", "")

	ms, err := strconv.ParseUint(raw, 16, 64)
	if err != nil {
		return time.Time{}, err
	}

	if ms > math.MaxInt64 {
		return time.Time{}, ErrBadTimestamp
	}

	return time.UnixMilli(int64(ms)).UTC(), nil
}

func main() {
	var v4s, v7s []string

	for range 5 {
		v4s = append(v4s, uuid.NewString())

		id, err := uuid.NewV7()
		if err != nil {
			fail(err)
		}

		v7s = append(v7s, id.String())
	}

	fmt.Println("== v4 (random) ==")

	for _, id := range v4s {
		fmt.Println(" ", id)
	}

	fmt.Println("== v7 (time-ordered) ==")

	for _, id := range v7s {
		fmt.Println(" ", id)
	}

	fmt.Println("v4 generated in sort order:", ordered(v4s))
	fmt.Println("v7 generated in sort order:", ordered(v7s))

	born, err := v7BornAt(v7s[0])
	if err != nil {
		fail(err)
	}

	fmt.Println("v7[0] born at:", born.Format(time.RFC3339))

	for i, id := range v7s {
		at, err := v7BornAt(id)
		if err != nil {
			fail(err)
		}

		fmt.Printf("v7[%d] %s born %s\n", i, id, at.Format("15:04:05.000"))
	}

	sorted := append([]string(nil), v7s...)
	sort.Strings(sorted)
	fmt.Println("v7 first bytes carry time (compare prefixes above).")
}
