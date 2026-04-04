package token

import (
	"testing"
	"time"
)

func TestGenerateAndValidateAccess(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"

	tokenStr, err := GenerateAccess(userID, secret, 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccess failed: %v", err)
	}

	got, err := ValidateAccess(tokenStr, secret)
	if err != nil {
		t.Fatalf("ValidateAccess failed: %v", err)
	}

	if got != userID {
		t.Errorf("want userID %q, got %q", userID, got)
	}
}

func TestValidateAccess_ExpiredToken(t *testing.T) {
	secret := "test-secret"

	// 이미 만료된 토큰 생성
	tokenStr, err := GenerateAccess("user-123", secret, -1*time.Second)
	if err != nil {
		t.Fatalf("GenerateAccess failed: %v", err)
	}

	_, err = ValidateAccess(tokenStr, secret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestValidateAccess_WrongSecret(t *testing.T) {
	tokenStr, err := GenerateAccess("user-123", "secret-a", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccess failed: %v", err)
	}

	_, err = ValidateAccess(tokenStr, "secret-b")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestGenerateRefresh(t *testing.T) {
	plain1, hash1, err := GenerateRefresh()
	if err != nil {
		t.Fatalf("GenerateRefresh failed: %v", err)
	}

	plain2, hash2, err := GenerateRefresh()
	if err != nil {
		t.Fatalf("GenerateRefresh failed: %v", err)
	}

	// 매번 다른 토큰이 생성되어야 한다
	if plain1 == plain2 {
		t.Error("expected different plain tokens")
	}
	if hash1 == hash2 {
		t.Error("expected different token hashes")
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token"

	// 같은 입력이면 같은 해시
	h1 := HashToken(token)
	h2 := HashToken(token)
	if h1 != h2 {
		t.Error("same input should produce same hash")
	}

	// 다른 입력이면 다른 해시
	h3 := HashToken("different-token")
	if h1 == h3 {
		t.Error("different input should produce different hash")
	}
}
