package main

import (
	"fmt"
	"go-lab/security/hashing/xcrypto"
)

func main() {
	s := "j9w3ykg7w2vw3ly60vxlq1ojytwbm6wt"

	hash1, _ := xcrypto.HashSHA256(s)
	fmt.Println("sha256:", hash1)

	hash2, _ := xcrypto.HashArgon2(s)
	fmt.Println("argon2:", hash2)

	h3, _ := xcrypto.BcryptHash(s)
	fmt.Println("bcrypt:", h3)

	blake3 := xcrypto.Blake3Hash(s)
	fmt.Println("blake3:", blake3)

	scrypt := xcrypto.ScryptHash(s)
	fmt.Println("scrypt:", scrypt)

	xxhash := xcrypto.XxHashHash(s)
	fmt.Println("xxhash:", xxhash)

	murmur3 := xcrypto.MurMur3Hash(s)
	fmt.Println("murmur3:", murmur3)

	farm := xcrypto.FarmHash(s)
	fmt.Println("farm:", farm)
}
