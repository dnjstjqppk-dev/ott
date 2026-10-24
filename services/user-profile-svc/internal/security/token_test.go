package security

import (
	"testing"
	"time"
)

func TestTokenIssueAndVerify(t *testing.T) {
	mgr := NewTokenManager("strict_production_hmac_secret_2026", 15*time.Minute)
	token, ttl, err := mgr.IssueAccessToken("u-1", "p-1", "d-1", "USER", "AGE_19", "KR")
	if err != nil {
		t.Fatalf("expected no error issuing token, got %v", err)
	}
	if ttl != 900 {
		t.Fatalf("expected 900s ttl, got %d", ttl)
	}

	claims, err := mgr.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("expected valid token verification, got %v", err)
	}
	if claims.UserID != "u-1" || claims.ProfileID != "p-1" || claims.MaxMaturity != "AGE_19" {
		t.Fatalf("unexpected claims payload: %+v", claims)
	}
}

func TestRefreshTokenPairAndBcrypt(t *testing.T) {
	raw, hashHex, err := GenerateRefreshTokenPair()
	if err != nil {
		t.Fatalf("unexpected error generating refresh token: %v", err)
	}
	if len(raw) != 64 || len(hashHex) != 64 {
		t.Fatalf("expected 64-char hex strings, got %d and %d", len(raw), len(hashHex))
	}
	if HashSHA256(raw) != hashHex {
		t.Fatalf("sha256 mismatch between raw and computed hash")
	}

	hashedPin, err := HashSecretBcrypt("2580")
	if err != nil {
		t.Fatalf("bcrypt hash error: %v", err)
	}
	if !CompareBcrypt(hashedPin, "2580") {
		t.Fatalf("expected bcrypt comparison to succeed")
	}
	if CompareBcrypt(hashedPin, "9999") {
		t.Fatalf("expected wrong pin to fail bcrypt check")
	}
}

