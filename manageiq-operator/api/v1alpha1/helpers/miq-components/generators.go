package miqtools

import (
	"crypto/rand"
	"encoding/base64"
)

func randomBytes(n int) []byte {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err) // out of randomness, should never happen
	}
	return buf
}

func generateEncryptionKey() string {
	// 32 random bytes encoded as URL-safe base64 (no padding), producing 43 chars from the alphabet [A-Za-z0-9_-]
	return base64.RawURLEncoding.EncodeToString(randomBytes(32))
}

func generatePassword() string {
	// 12 random bytes encoded as URL-safe base64 (no padding), producing 16 chars from the alphabet [A-Za-z0-9_-]
	return base64.RawURLEncoding.EncodeToString(randomBytes(12))
}
