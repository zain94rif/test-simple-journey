package service

import (
	"context"
	"testing"
	"time"

	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/domain"
)

func TestRegisterHashesPasswordAndForcesRequester(t *testing.T) {
	users := newFakeUserRepo()
	manager := auth.NewManager("01234567890123456789012345678901", time.Hour)
	svc := NewAuthService(users, manager, 10)

	created, err := svc.Register(context.Background(), RegisterInput{
		Name:     "Budi Santoso",
		Email:    "Budi@Example.com",
		Password: "Rahasia123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Role != domain.RoleRequester {
		t.Fatalf("expected requester, got %s", created.Role)
	}
	if created.Email != "budi@example.com" {
		t.Fatalf("expected lowercase email, got %s", created.Email)
	}

	stored := users.users[created.ID]
	if stored.PasswordHash == "Rahasia123" || !auth.VerifyPassword(stored.PasswordHash, "Rahasia123") {
		t.Fatalf("password was not hashed correctly")
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	users := newFakeUserRepo()
	manager := auth.NewManager("01234567890123456789012345678901", time.Hour)
	svc := NewAuthService(users, manager, 10)

	_, err := svc.Register(context.Background(), RegisterInput{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "abcdefgh",
	})
	requireAppCode(t, err, "VALIDATION_ERROR")
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("Rahasia123", 10)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	users := newFakeUserRepo(domain.User{
		ID: 1, Name: "Budi", Email: "budi@example.com", PasswordHash: hash, Role: domain.RoleRequester,
	})
	manager := auth.NewManager("01234567890123456789012345678901", time.Hour)
	svc := NewAuthService(users, manager, 10)

	_, err = svc.Login(context.Background(), LoginInput{Email: "budi@example.com", Password: "salah"})
	requireAppCode(t, err, "INVALID_CREDENTIALS")
}

func TestLoginReturnsParseableToken(t *testing.T) {
	hash, err := auth.HashPassword("Rahasia123", 10)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	users := newFakeUserRepo(domain.User{
		ID: 2, Name: "Agent", Email: "agent@example.com", PasswordHash: hash, Role: domain.RoleAgent,
	})
	manager := auth.NewManager("01234567890123456789012345678901", time.Hour)
	svc := NewAuthService(users, manager, 10)

	result, err := svc.Login(context.Background(), LoginInput{Email: "agent@example.com", Password: "Rahasia123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	userID, role, err := manager.Parse(result.Token)
	if err != nil {
		t.Fatalf("token should be valid: %v", err)
	}
	if userID != 2 || role != domain.RoleAgent {
		t.Fatalf("unexpected claims: id=%d role=%s", userID, role)
	}
}
