package middleware

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Skuyno/room-booking/internal/auth"
)

func TestAuthMiddlewarePublicPathBypassesToken(t *testing.T) {
	t.Parallel()

	called := false
	handler := Auth(auth.NewJWTManager("secret"))(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		called = true
		w.WriteHeader(stdhttp.StatusOK)
	}))

	req := httptest.NewRequest(stdhttp.MethodGet, "/_info", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestAuthMiddlewareRejectsMissingBearerToken(t *testing.T) {
	t.Parallel()

	handler := Auth(auth.NewJWTManager("secret"))(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		t.Fatal("next handler must not be called")
	}))

	req := httptest.NewRequest(stdhttp.MethodGet, "/rooms/list", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if body["error"]["code"] != "UNAUTHORIZED" {
		t.Fatalf("code = %s, want UNAUTHORIZED", body["error"]["code"])
	}
}

func TestAuthMiddlewareStoresClaimsInContext(t *testing.T) {
	t.Parallel()

	jwtManager := auth.NewJWTManager("secret")
	token, err := jwtManager.Generate("11111111-1111-1111-1111-111111111111", "user", time.Hour)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var claimsFound bool
	handler := Auth(jwtManager)(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok {
			t.Fatal("claims not found in context")
		}
		claimsFound = true
		if claims.UserID != "11111111-1111-1111-1111-111111111111" || claims.Role != "user" {
			t.Fatalf("claims = %+v", claims)
		}
		w.WriteHeader(stdhttp.StatusOK)
	}))

	req := httptest.NewRequest(stdhttp.MethodGet, "/rooms/list", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !claimsFound {
		t.Fatal("claims were not checked")
	}
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if _, ok := GetClaims(context.Background()); ok {
		t.Fatal("GetClaims() on empty context must return ok=false")
	}
}
