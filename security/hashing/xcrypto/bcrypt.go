package xcrypto

import "golang.org/x/crypto/bcrypt"

func BcryptHash(password string) (string, error) {
	// HighCost tradeoff is CPU and RAM
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func BcryptCheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
