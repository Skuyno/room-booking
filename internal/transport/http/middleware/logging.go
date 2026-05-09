// Package middleware содержит HTTP-middleware приложения.
package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// statusRecorder перехватывает WriteHeader, чтобы запомнить статус ответа
// для логирования. Дефолтно считается StatusOK, если хендлер не вызвал
// WriteHeader явно.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader сохраняет код и проксирует вызов в обёрнутый ResponseWriter.
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging логирует базовую информацию по каждому запросу: IP, метод, путь,
// статус, длительность и user-agent. Все строки санитизируются от \r\n,
// чтобы исключить log injection через user-controlled-поля.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		ip := sanitizeLogValue(clientIP(r))
		method := sanitizeLogValue(r.Method)
		path := sanitizeLogValue(r.URL.Path)
		ua := sanitizeLogValue(r.UserAgent())

		//nolint:gosec // values are sanitized with sanitizeLogValue before logging
		log.Printf(
			"ip=%s method=%s path=%s status=%d duration=%s ua=%q",
			ip,
			method,
			path,
			rec.status,
			time.Since(start),
			ua,
		)
	})
}

// clientIP возвращает best-effort IP клиента: сначала смотрит
// X-Forwarded-For (первый адрес — оригинальный клиент за прокси), потом
// X-Real-IP, потом RemoteAddr.
//
// Доверять заголовкам можно только если перед сервисом стоит trusted-прокси,
// который их выставляет. В открытом интернете эти заголовки спуфятся.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

// sanitizeLogValue убирает символы переноса строки, чтобы юзер не мог
// подделать строку лога через свой user-agent или путь.
func sanitizeLogValue(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}
