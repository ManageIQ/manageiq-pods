package miqtools

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"regexp"
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
	for {
		buf := randomBytes(8)
		password := hex.EncodeToString(buf)
		if match, err := regexp.MatchString(`\D+`, password); err == nil && match {
			// Only return if a letter is included.
			// Password decryption can fail if the database password is all numbers because ruby will read it as an integer instead of a string.
			return password
		}
	}
}
