package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("Rahasia123", 10)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Rahasia123" {
		t.Fatal("password must not be stored in plaintext")
	}
	if !VerifyPassword(hash, "Rahasia123") {
		t.Fatal("correct password should verify")
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("Rahasia123", 10)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if VerifyPassword(hash, "Rahasia124") {
		t.Fatal("wrong password must not verify")
	}
}
