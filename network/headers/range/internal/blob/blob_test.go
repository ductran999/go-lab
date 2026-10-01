package blob

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"testing"
)

var digestPattern = regexp.MustCompile(`^sha-256=:[A-Za-z0-9+/]+={0,2}:$`)

func TestDigestFormat(t *testing.T) {
	t.Parallel()

	_, digest := Build(1)

	if !digestPattern.MatchString(digest) {
		t.Fatalf("digest %q is not RFC 9530 sha-256=:...: (no trailing junk)", digest)
	}
}

func TestDigestMatchesBlob(t *testing.T) {
	t.Parallel()

	blob, digest := Build(1)

	inner := digest[len("sha-256=:") : len(digest)-1]

	sum := sha256.Sum256(blob)

	if got := base64.StdEncoding.EncodeToString(sum[:]); got != inner {
		t.Fatalf("digest covers different bytes than served blob")
	}
}
