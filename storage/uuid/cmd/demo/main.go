// Command demo prints v4 vs v7 side by side: structure, sort order,
// and what each leaks. Usage: go run ./cmd/demo
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/google/uuid"
)

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

	sorted := append([]string(nil), v7s...)
	sort.Strings(sorted)
	fmt.Println("v7 first bytes carry time (compare prefixes above).")
}
