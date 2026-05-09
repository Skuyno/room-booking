// Package auth содержит низкоуровневые крипто-утилиты: подпись и проверку
// JWT-токенов и хэширование/сверку паролей через bcrypt. Этим пакетом
// пользуются auth_service (выдача токена при логине) и middleware/auth
// (валидация токена на входящем запросе).
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — содержимое JWT-токена. Включает стандартные поля (exp, iat) и
// два кастомных: UserID и Role. Эти два значения после валидации токена
// кладутся в request context и используются хендлерами для авторизации.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager подписывает и валидирует JWT-токены симметричным ключом (HS256).
// Создаётся один раз в composition root (cmd/api/main.go) и инжектится в
// AuthService и middleware. Секрет хранится в []byte, чтобы не пересоздавать
// его на каждый вызов.
type JWTManager struct {
	secret []byte
}

// NewJWTManager создаёт менеджер с заданным симметричным секретом.
func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
	}
}

// Generate подписывает новый токен с UserID, Role и сроком жизни ttl.
// Возвращает строковое представление токена (готовое для Authorization-заголовка).
func (m *JWTManager) Generate(userID, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse валидирует подпись и срок жизни токена, возвращает разобранные claims.
// Явно проверяет alg=HS256 — это страховка от подделки токенов с alg=none.
func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("token invalid")
	}

	return claims, nil
}
