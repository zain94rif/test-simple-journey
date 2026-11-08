package auth

import (
	"errors"
	"testing"
	"time"

	"helpdesk/backend/internal/domain"
)

const testSecret = "01234567890123456789012345678901"

func TestGenerateAndParseToken(t *testing.T) {
	manager := NewManager(testSecret, time.Hour)
	token, expiresAt, err := manager.Generate(42, domain.RoleAgent)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !expiresAt.After(time.Now()) {
		t.Fatal("expiry should be in the future")
	}

	userID, role, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if userID != 42 || role != domain.RoleAgent {
		t.Fatalf("unexpected claims: id=%d role=%s", userID, role)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	issuer := NewManager(testSecret, time.Hour)
	token, _, err := issuer.Generate(1, domain.RoleAdmin)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	verifier := NewManager("abcdefghijklmnopqrstuvwxyz123456", time.Hour)
	if _, _, err := verifier.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	manager := NewManager(testSecret, -time.Minute)
	token, _, err := manager.Generate(1, domain.RoleAdmin)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, _, err := manager.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseRejectsMalformedToken(t *testing.T) {
	manager := NewManager(testSecret, time.Hour)
	if _, _, err := manager.Parse("not-a-jwt"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
