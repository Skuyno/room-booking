package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Skuyno/room-booking/internal/auth"
)

// contextKey — приватный тип для ключей контекста, чтобы случайно не
// пересечься с ключами из других пакетов.
type contextKey string

// ClaimsKey — ключ, под которым Auth-middleware кладёт *auth.Claims в
// request context. Хендлеры читают через GetClaims.
const ClaimsKey contextKey = "claims"

// Auth возвращает middleware, валидирующее JWT в заголовке
// Authorization: Bearer <token>. При успехе кладёт распаршенные claims в
// контекст; иначе — отвечает 401 и не передаёт запрос дальше.
//
// Публичные пути (/_info, /dummyLogin, /login, /register) пропускаются
// без токена. Список захардкожен в фабрике, чтобы middleware не лазил в
// конфиг каждый запрос.
func Auth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	publicPaths := map[string]struct{}{
		"/_info":      {},
		"/dummyLogin": {},
		"/login":      {},
		"/register":   {},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := publicPaths[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtManager.Parse(token)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims достаёт *auth.Claims из контекста запроса. Возвращает (nil, false),
// если middleware не положил claims — то есть запрос пришёл по публичному
// пути или не прошёл авторизацию.
func GetClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*auth.Claims)
	return claims, ok
}
