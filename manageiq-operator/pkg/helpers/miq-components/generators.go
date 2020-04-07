package miqtools

import (
	"crypto/rand"
	"crypto/sha256"
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
	sum := sha256.Sum256(randomBytes(32))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func generateDatabasePassword() string {
	buf := randomBytes(8)
	return hex.EncodeToString(buf)
}
