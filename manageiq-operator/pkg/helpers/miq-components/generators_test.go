package miqtools

import (
	"testing"
)

func TestRandomBytes(t *testing.T) {
	expected := 10
	bytes := randomBytes(expected)
	actual := len(bytes)

	if actual != expected {
		t.Fatalf(`Expected %d bytes, but got %d bytes.`, expected, actual)
	}
}
