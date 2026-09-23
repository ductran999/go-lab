package xcrypto

import (
	"fmt"

	"github.com/dgryski/go-farm"
)

func FarmHash(plaintext string) string {
	hash := farm.Hash64([]byte(plaintext))
	return fmt.Sprintf("%d", hash)
}
