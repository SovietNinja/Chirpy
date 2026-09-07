package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "correct-horse-battery-staple"

	token, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	gotID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	if gotID != userID {
		t.Errorf("ValidateJWT returned userID %v, want %v", gotID, userID)
	}
}

func TestValidateJWTExpired(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "correct-horse-battery-staple"

	token, err := MakeJWT(userID, tokenSecret, -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Error("ValidateJWT did not return an error for an expired token")
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "correct-horse-battery-staple"
	wrongSecret := "wrong-secret"

	token, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("ValidateJWT did not return an error for a token signed with the wrong secret")
	}
}
