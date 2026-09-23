package xcrypto

import (
	"crypto/sha3"
	"encoding/hex"
)

func SHA3Hash(plaintext string) string {
	input := "secure_data_to_hash"

	hash256 := sha3.Sum256([]byte(input))

	return hex.EncodeToString(hash256[:])
}
