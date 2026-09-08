package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		authValue string
		setHeader bool
		wantToken string
		wantErr   bool
	}{
		{name: "correct header", authValue: "Bearer 1234567890", setHeader: true, wantToken: "1234567890"},
		{name: "padded token", authValue: "Bearer   1234567890  ", setHeader: true, wantToken: "1234567890"},
		{name: "malformed prefix", authValue: "Bear 676767", setHeader: true, wantErr: true},
		{name: "only Bearer prefix", authValue: "Bearer", setHeader: true, wantErr: true},
		{name: "only Bearer prefix with space", authValue: "Bearer  ", setHeader: true, wantErr: true},
		{name: "no Authorization key", setHeader: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			if tt.setHeader {
				h.Set("Authorization", tt.authValue)
			} else {
				h.Set("Login", "")
			}

			token, err := GetBearerToken(h)

			if tt.wantErr {
				if err == nil {
					t.Error("GetBearerToken expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetBearerToken returned unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("GetBearerToken returned wrong token: got %q, want %q", token, tt.wantToken)
			}
		})
	}
}
