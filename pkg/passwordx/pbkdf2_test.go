package passwordx

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestVerify(t *testing.T) {
	const (
		rawPassword = "pass123456"
		salt        = "salt-value"
		iterations  = 10000
		keyLen      = 128
	)

	hash := pbkdf2.Key([]byte(rawPassword), []byte(salt), iterations, keyLen, sha512.New)
	encoded := hex.EncodeToString(hash)

	if !Verify(rawPassword, salt, encoded) {
		t.Fatal("Verify() = false, want true for matching password")
	}
	if Verify("wrong-password", salt, encoded) {
		t.Fatal("Verify() = true, want false for mismatched password")
	}
}

func TestEncodeAndVerify(t *testing.T) {
	salt, encoded := Encode("safe-password")

	if salt == "" {
		t.Fatal("Encode() salt is empty")
	}
	if encoded == "" {
		t.Fatal("Encode() encoded password is empty")
	}
	if !Verify("safe-password", salt, encoded) {
		t.Fatal("Verify() = false, want true for encoded password")
	}
}
