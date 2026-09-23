package xcrypto

import (
	"encoding/hex"

	"golang.org/x/crypto/scrypt"
)

func ScryptHash(plaintext string) string {
	salt := []byte("random_salt_here")

	// N=16384 (CPU/Memory cost), r=8 (Block size), p=1 (Parallelism)
	hash, _ := scrypt.Key([]byte(plaintext), salt, 16384, 8, 1, 32)

	return hex.EncodeToString(hash)
}
