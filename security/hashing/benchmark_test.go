package main

import (
	"go-lab/security/hashing/xcrypto"
	"testing"
)

var plaintext = "j9w3ykg7w2vw3ly60vxlq1ojytwbm6wt"

// Slow Hash

func BenchmarkArgon2(b *testing.B) {
	b.ResetTimer()

	for b.Loop() {
		_, _ = xcrypto.HashArgon2(plaintext)
	}
}

func BenchmarkBcrypt(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_, _ = xcrypto.BcryptHash(plaintext)
	}
}

func BenchmarkScrypt(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_ = xcrypto.ScryptHash(plaintext)
	}
}

// Balance

func BenchmarkSHA256(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_, _ = xcrypto.HashSHA256(plaintext)
	}
}

func BenchmarkSHA3(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_ = xcrypto.SHA3Hash(plaintext)
	}
}

func BenchmarkBLAKE3(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_ = xcrypto.Blake3Hash(plaintext)
	}
}

// Super Fast

func BenchmarkXxHash(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_ = xcrypto.XxHashHash(plaintext)
	}
}

func BenchmarkFarm(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_ = xcrypto.FarmHash(plaintext)
	}
}
func BenchmarkMurMur3(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		_ = xcrypto.MurMur3Hash(plaintext)
	}
}
