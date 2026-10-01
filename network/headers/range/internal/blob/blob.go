// Package blob builds the deterministic readable text served by
// the range lab, plus its RFC 9530 Content-Digest value.
package blob

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Build returns targetMB megabytes of numbered lines and the
// Content-Digest header value (`sha-256=:...:`) covering them.
func Build(targetMB int) ([]byte, string) {
	target := targetMB << 20

	var buf bytes.Buffer

	for i := 0; buf.Len() < target; i++ {
		fmt.Fprintf(&buf, "line-%06d the quick brown fox jumps over the lazy dog\n", i)
	}

	blob := buf.Bytes()
	sum := sha256.Sum256(blob)

	return blob, "sha-256=:" + base64.StdEncoding.EncodeToString(sum[:]) + ":"
}
