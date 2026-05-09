// Package config загружает конфигурацию приложения из переменных окружения.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config агрегирует все runtime-настройки приложения.
//
//   - HTTPPort  — порт, на котором слушает HTTP-сервер (например, "8080").
//   - DBDSN     — DSN для PostgreSQL (postgres://user:pass@host:port/db?sslmode=disable).
//   - JWTSecret — симметричный ключ для подписи JWT-токенов (HS256).
type Config struct {
	HTTPPort  string
	DBDSN     string
	JWTSecret string
}

// ErrMissingEnv возвращается из Load, если хотя бы одна обязательная
// переменная окружения не задана. Сообщение содержит список всех
// недостающих переменных, чтобы поднять сразу все за один деплой.
var ErrMissingEnv = errors.New("missing required environment variables")

// Load читает конфигурацию из переменных окружения и валидирует её. Все три
// переменные обязательны — при отсутствии любой возвращается ошибка с
// перечислением всех недостающих, чтобы main мог упасть с понятным
// сообщением, а не молча подписывать токены пустым ключом.
func Load() (Config, error) {
	cfg := Config{
		HTTPPort:  strings.TrimSpace(os.Getenv("HTTP_PORT")),
		DBDSN:     strings.TrimSpace(os.Getenv("DB_DSN")),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	var missing []string
	if cfg.HTTPPort == "" {
		missing = append(missing, "HTTP_PORT")
	}
	if cfg.DBDSN == "" {
		missing = append(missing, "DB_DSN")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("%w: %s", ErrMissingEnv, strings.Join(missing, ", "))
	}

	return cfg, nil
}
