// Command hashproof shows block hashing equals whole hashing:
// same bytes in the same order → same digest, whatever the cut.
//
//	go run ./cmd/hashproof -file ./tmp/lines.txt
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "hashproof:", err)

	os.Exit(1)
}

func main() {
	file := flag.String("file", "./tmp/lines.txt", "file to hash")

	flag.Parse()

	data, err := os.ReadFile(*file)
	if err != nil {
		fail(err)
	}

	whole := sha256.Sum256(data)

	// Same bytes, awkward cuts: 7, 65537, rest.
	cuts := []int{7, 65537}

	stream := sha256.New()
	pos := 0

	for _, c := range cuts {
		if pos+c > len(data) {
			break
		}

		_, _ = stream.Write(data[pos : pos+c])
		pos += c
	}

	_, _ = stream.Write(data[pos:])

	fmt.Printf("whole:  %s\nstream: %s\n", hex.EncodeToString(whole[:]), hex.EncodeToString(stream.Sum(nil)))
	fmt.Printf("equal: %t\n", hex.EncodeToString(whole[:]) == hex.EncodeToString(stream.Sum(nil)))
}
