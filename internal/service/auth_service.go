package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/auth"
	"github.com/Skuyno/room-booking/internal/domain"
)

// Фиксированные UUID для тестовой ручки /dummyLogin. Те же значения
// засеяны в migrations/003_seed_dummy_data.sql, поэтому JWT, выданный
// здесь, валиден против реальных строк в users.
const (
	adminUserID  = "11111111-1111-1111-1111-111111111111"
	normalUserID = "22222222-2222-2222-2222-222222222222"
)

// AuthService отвечает за регистрацию, логин и тестовую авторизацию через
// /dummyLogin. Подписывает JWT через JWTManager и хранит/ищет пользователей
// через UserRepository.
type AuthService struct {
	jwt      *auth.JWTManager
	userRepo UserRepository
}

// NewAuthService собирает сервис из готового JWT-менеджера и репозитория.
func NewAuthService(jwt *auth.JWTManager, userRepo UserRepository) *AuthService {
	return &AuthService{
		jwt:      jwt,
		userRepo: userRepo,
	}
}

// DummyLogin выдаёт тестовый JWT с фиксированным UUID, соответствующим
// переданной роли. Для роли "admin" — adminUserID, для "user" — normalUserID.
// Для любого другого значения возвращает domain.ErrInvalidRequest.
func (s *AuthService) DummyLogin(_ context.Context, role string) (string, error) {
	var userID string

	switch role {
	case domain.RoleAdmin:
		userID = adminUserID
	case domain.RoleUser:
		userID = normalUserID
	default:
		return "", domain.ErrInvalidRequest
	}

	return s.jwt.Generate(userID, role, 24*time.Hour)
}

// Register создаёт нового пользователя с указанной ролью. Email
// нормализуется (lower + trim) и валидируется через mail.ParseAddress.
// Пароль хэшируется через bcrypt. Возвращает domain.ErrEmailTaken, если
// email уже занят.
func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
	role string,
) (domain.User, error) {
	email = normalizeEmail(email)
	if err := validateCredentials(email, password); err != nil {
		return domain.User{}, err
	}

	roleID := domain.RoleIDByCode(role)
	if roleID == 0 {
		return domain.User{}, domain.ErrInvalidRequest
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:           uuid.New(),
		Email:        email,
		RoleID:       roleID,
		Role:         role,
		PasswordHash: passwordHash,
	}

	return s.userRepo.Create(ctx, user)
}

// Login сверяет пароль и выдаёт JWT. ВСЕ ошибки валидации, отсутствия юзера
// и неверного пароля сводятся к ErrInvalidCredentials, чтобы наружу не
// сигналить о существовании email (защита от user enumeration).
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if err := validateCredentials(email, password); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if user.PasswordHash == "" {
		return "", domain.ErrInvalidCredentials
	}

	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.jwt.Generate(user.ID.String(), user.Role, 24*time.Hour)
}

// normalizeEmail приводит email к нижнему регистру и убирает пробелы.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// validateCredentials проверяет базовый формат email и непустой пароль.
// Возвращает domain.ErrInvalidRequest при любой проблеме.
func validateCredentials(email, password string) error {
	if email == "" || password == "" {
		return domain.ErrInvalidRequest
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return domain.ErrInvalidRequest
	}

	if strings.TrimSpace(password) == "" {
		return domain.ErrInvalidRequest
	}

	return nil
}
