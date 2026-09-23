package xcrypto

import (
	"crypto/sha256"
	"encoding/hex"
)

var hasherSHA256 = sha256.New()

func HashSHA256(plaintext string) (string, error) {
	salt := "internal_system_secret_key"
	_, err := hasherSHA256.Write([]byte(plaintext + salt))
	if err != nil {
		return "", err
	}

	hashBytes := hasherSHA256.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString, nil
}
