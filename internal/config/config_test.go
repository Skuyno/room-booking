package config

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadSuccess(t *testing.T) {
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DB_DSN", "postgres://test")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("HTTPPort = %s, want 8080", cfg.HTTPPort)
	}
	if cfg.DBDSN != "postgres://test" {
		t.Fatalf("DBDSN = %s, want postgres://test", cfg.DBDSN)
	}
	if cfg.JWTSecret != "secret" {
		t.Fatalf("JWTSecret = %s, want secret", cfg.JWTSecret)
	}
}

func TestLoadFailsWithoutRequiredEnv(t *testing.T) {
	t.Setenv("HTTP_PORT", "")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	if !errors.Is(err, ErrMissingEnv) {
		t.Fatalf("Load() error = %v, want %v", err, ErrMissingEnv)
	}
}

func TestLoadReportsAllMissingVars(t *testing.T) {
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DB_DSN", "")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing vars")
	}
	msg := err.Error()
	if !strings.Contains(msg, "DB_DSN") || !strings.Contains(msg, "JWT_SECRET") {
		t.Fatalf("error message = %q, want to mention both DB_DSN and JWT_SECRET", msg)
	}
	if strings.Contains(msg, "HTTP_PORT") {
		t.Fatalf("error message = %q, must not mention HTTP_PORT", msg)
	}
}
