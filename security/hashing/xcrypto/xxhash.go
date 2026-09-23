package xcrypto

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
)

func XxHashHash(plaintext string) string {
	hash := xxhash.Sum64([]byte(plaintext))
	return fmt.Sprintf("%d", hash)
}
