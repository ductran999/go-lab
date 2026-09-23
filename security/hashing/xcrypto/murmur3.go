package xcrypto

import (
	"fmt"

	"github.com/spaolacci/murmur3"
)

func MurMur3Hash(plaintext string) string {
	h1 := murmur3.Sum64([]byte(plaintext))
	return fmt.Sprintf("%d", h1)
}
