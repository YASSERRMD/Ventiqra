package auth

import (
	"errors"
	"testing"
	"time"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "supersecret" {
		t.Fatal("hash empty or equal to plaintext")
	}
	if err := CheckPassword(hash, "supersecret"); err != nil {
		t.Errorf("CheckPassword correct: %v", err)
	}
	if err := CheckPassword(hash, "wrong"); err == nil {
		t.Error("CheckPassword wrong: expected error")
	}
}

func TestHashPasswordRejectsShort(t *testing.T) {
	if _, err := HashPassword("short"); err == nil {
		t.Error("expected error for short password")
	}
}

func TestTokenIssueAndParse(t *testing.T) {
	tm, err := NewTokenManager("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewTokenManager: %v", err)
	}

	token, exp, err := tm.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if !exp.After(time.Now()) {
		t.Error("expiry not in the future")
	}

	claims, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("uid = %q, want user-123", claims.UserID)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	tm1, _ := NewTokenManager("secret-one", time.Hour)
	tm2, _ := NewTokenManager("secret-two", time.Hour)

	token, _, _ := tm1.Issue("user-1")
	if _, err := tm2.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseRejectsExpired(t *testing.T) {
	tm, _ := NewTokenManager("secret", 5*time.Millisecond)
	token, _, _ := tm.Issue("user-1")
	time.Sleep(20 * time.Millisecond)
	if _, err := tm.Parse(token); !errors.Is(err, ErrExpiredToken) {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestNewTokenManagerRequiresSecret(t *testing.T) {
	if _, err := NewTokenManager("", time.Hour); err == nil {
		t.Error("expected error for empty secret")
	}
}
