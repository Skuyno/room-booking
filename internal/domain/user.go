package domain

import (
	"time"

	"github.com/google/uuid"
)

// Строковые коды ролей. Используются в JWT-claims, в API и в логике сервисов.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Числовые ID ролей. Хранятся в users.role_id и в таблице-справочнике
// user_roles. Существуют, чтобы FK был на компактный SMALLINT, а не на TEXT.
const (
	RoleIDAdmin int16 = 1
	RoleIDUser  int16 = 2
)

// User — пользователь системы. PasswordHash пуст для seed-юзеров под
// /dummyLogin (они авторизуются по фиксированному UUID, а не паролю).
type User struct {
	ID           uuid.UUID
	Email        string
	RoleID       int16
	Role         string
	PasswordHash string
	CreatedAt    time.Time
}

// RoleCodeByID конвертирует числовой ID роли в строковой код.
// Возвращает пустую строку для неизвестного ID.
func RoleCodeByID(id int16) string {
	switch id {
	case RoleIDAdmin:
		return RoleAdmin
	case RoleIDUser:
		return RoleUser
	default:
		return ""
	}
}

// RoleIDByCode конвертирует строковой код роли в числовой ID.
// Возвращает 0 для неизвестного кода — сервисы трактуют это как невалидную роль.
func RoleIDByCode(code string) int16 {
	switch code {
	case RoleAdmin:
		return RoleIDAdmin
	case RoleUser:
		return RoleIDUser
	default:
		return 0
	}
}
