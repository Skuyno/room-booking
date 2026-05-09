package middleware

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 203.0.113.11")
	if ip := clientIP(req); ip != "203.0.113.10" {
		t.Fatalf("clientIP() = %s, want 203.0.113.10", ip)
	}

	req = httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.7")
	if ip := clientIP(req); ip != "198.51.100.7" {
		t.Fatalf("clientIP() = %s, want 198.51.100.7", ip)
	}

	req = httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.5:1234"
	if ip := clientIP(req); ip != "192.0.2.5" {
		t.Fatalf("clientIP() = %s, want 192.0.2.5", ip)
	}
}

func TestLoggingMiddlewarePassesStatusCode(t *testing.T) {
	t.Parallel()

	handler := Logging(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusCreated)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(stdhttp.MethodPost, "/rooms/create", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
}
