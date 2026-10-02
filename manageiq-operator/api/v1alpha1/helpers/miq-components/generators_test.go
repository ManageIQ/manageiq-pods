package miqtools

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	t.Run("produces a 16 character string", func(t *testing.T) {
		password := generatePassword()
		if len(password) != 16 {
			t.Errorf("expected length 16, got %d (%q)", len(password), password)
		}
	})

	t.Run("is valid raw URL-safe base64", func(t *testing.T) {
		password := generatePassword()
		if _, err := base64.RawURLEncoding.DecodeString(password); err != nil {
			t.Errorf("expected valid raw URL base64, got %q: %v", password, err)
		}
	})

	t.Run("contains only URL-safe base64 characters (no +, /, or =)", func(t *testing.T) {
		const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
		password := generatePassword()
		for _, c := range password {
			if !strings.ContainsRune(alphabet, c) {
				t.Errorf("unexpected character %q in password %q", c, password)
			}
		}
	})

	t.Run("produces unique values", func(t *testing.T) {
		seen := make(map[string]struct{})
		for i := 0; i < 100; i++ {
			pw := generatePassword()
			if _, exists := seen[pw]; exists {
				t.Errorf("duplicate password generated: %q", pw)
			}
			seen[pw] = struct{}{}
		}
	})
}

func TestGenerateEncryptionKey(t *testing.T) {
	t.Run("is valid standard base64 encoding 32 bytes of data", func(t *testing.T) {
		key := generateEncryptionKey()
		decoded, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			t.Errorf("expected valid standard base64, got %q: %v", key, err)
		}
		if len(decoded) != 32 {
			t.Errorf("expected 32 decoded bytes, got %d", len(decoded))
		}
	})

	t.Run("produces unique values", func(t *testing.T) {
		seen := make(map[string]struct{})
		for i := 0; i < 100; i++ {
			key := generateEncryptionKey()
			if _, exists := seen[key]; exists {
				t.Errorf("duplicate key generated: %q", key)
			}
			seen[key] = struct{}{}
		}
	})
}
