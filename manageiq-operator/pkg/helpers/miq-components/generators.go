package miqtools

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
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
	return base64.StdEncoding.EncodeToString(randomBytes(32))
}

func generatePassword() string {
	buf := randomBytes(8)
	return hex.EncodeToString(buf)
}
