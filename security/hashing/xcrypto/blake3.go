package xcrypto

import (
	"encoding/hex"

	"github.com/zeebo/blake3"
)

func Blake3Hash(plaintext string) string {
	key := []byte("a-very-secret-32-byte-long-key..")
	hasher := blake3.NewDeriveKey(string(key))
	hasher.Write([]byte(plaintext))
	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes)
}
