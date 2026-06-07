package auth

import (
	"testing"
	"time"
)

func TestGenerateUserTokenAndParseUserID(t *testing.T) {
	issuedAt := time.Now()

	token, err := GenerateUserToken("secret", issuedAt, 3600, 99)
	if err != nil {
		t.Fatalf("GenerateUserToken() error = %v", err)
	}

	userID, err := ParseUserID(token, "secret")
	if err != nil {
		t.Fatalf("ParseUserID() error = %v", err)
	}
	if userID != 99 {
		t.Fatalf("ParseUserID() userID = %d, want 99", userID)
	}
}

func TestParseUserIDRejectsExpiredToken(t *testing.T) {
	token, err := GenerateUserToken("secret", time.Now().Add(-2*time.Hour), 60, 99)
	if err != nil {
		t.Fatalf("GenerateUserToken() error = %v", err)
	}

	if _, err = ParseUserID(token, "secret"); err == nil {
		t.Fatal("ParseUserID() error = nil, want expired-token error")
	}
}

func TestParseUserIDRejectsWrongSecret(t *testing.T) {
	token, err := GenerateUserToken("secret", time.Now(), 3600, 99)
	if err != nil {
		t.Fatalf("GenerateUserToken() error = %v", err)
	}

	if _, err = ParseUserID(token, "another-secret"); err == nil {
		t.Fatal("ParseUserID() error = nil, want signature error")
	}
}
