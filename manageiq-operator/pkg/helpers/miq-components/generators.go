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
	buf := randomBytes(32)

	h := sha256.New()
	h.Write(buf)

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func generatePassword() string {
	buf := randomBytes(8)
	return hex.EncodeToString(buf)
}
