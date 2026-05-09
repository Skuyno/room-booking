package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManagerGenerateAndParse(t *testing.T) {
	t.Parallel()

	manager := NewJWTManager("secret")

	token, err := manager.Generate("user-1", "user", time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("claims.UserID = %s, want user-1", claims.UserID)
	}
	if claims.Role != "user" {
		t.Fatalf("claims.Role = %s, want user", claims.Role)
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		t.Fatal("expected registered claims to be filled")
	}
}

func TestJWTManagerParseRejectsWrongSigningMethod(t *testing.T) {
	t.Parallel()

	token := jwt.NewWithClaims(jwt.SigningMethodHS384, Claims{UserID: "user-1", Role: "user"})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = NewJWTManager("secret").Parse(tokenString)
	if err == nil || !strings.Contains(err.Error(), "unexpected signing method") {
		t.Fatalf("Parse() error = %v, want unexpected signing method", err)
	}
}
