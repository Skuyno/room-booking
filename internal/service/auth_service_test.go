package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Skuyno/room-booking/internal/auth"
	"github.com/Skuyno/room-booking/internal/domain"
)

func TestAuthServiceDummyLoginReturnsFixedClaims(t *testing.T) {
	t.Parallel()

	jwtManager := auth.NewJWTManager("test-secret")
	svc := NewAuthService(jwtManager, userRepoStub{})

	cases := []struct {
		name           string
		role           string
		expectedUserID string
	}{
		{name: "admin", role: domain.RoleAdmin, expectedUserID: "11111111-1111-1111-1111-111111111111"},
		{name: "user", role: domain.RoleUser, expectedUserID: "22222222-2222-2222-2222-222222222222"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := svc.DummyLogin(context.Background(), tc.role)
			if err != nil {
				t.Fatalf("DummyLogin() error = %v", err)
			}

			claims, err := jwtManager.Parse(token)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if claims.UserID != tc.expectedUserID {
				t.Fatalf("claims.UserID = %s, want %s", claims.UserID, tc.expectedUserID)
			}
			if claims.Role != tc.role {
				t.Fatalf("claims.Role = %s, want %s", claims.Role, tc.role)
			}
		})
	}
}

func TestAuthServiceDummyLoginRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(auth.NewJWTManager("test-secret"), userRepoStub{})

	_, err := svc.DummyLogin(context.Background(), "guest")
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("DummyLogin() error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestAuthServiceRegisterNormalizesEmailAndHashesPassword(t *testing.T) {
	t.Parallel()

	var created domain.User
	repo := userRepoStub{
		createFn: func(_ context.Context, user domain.User) (domain.User, error) {
			created = user
			return user, nil
		},
	}

	svc := NewAuthService(auth.NewJWTManager("test-secret"), repo)

	user, err := svc.Register(context.Background(), " USER@Example.com ", "secret-pass", domain.RoleUser)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if created.Email != "user@example.com" {
		t.Fatalf("created.Email = %s, want user@example.com", created.Email)
	}
	if created.RoleID != domain.RoleIDUser {
		t.Fatalf("created.RoleID = %d, want %d", created.RoleID, domain.RoleIDUser)
	}
	if created.PasswordHash == "" || created.PasswordHash == "secret-pass" {
		t.Fatalf("created.PasswordHash must be set and hashed, got %q", created.PasswordHash)
	}
	if user.Email != created.Email {
		t.Fatalf("returned user email = %s, want %s", user.Email, created.Email)
	}
}

func TestAuthServiceLoginMapsUserNotFoundToInvalidCredentials(t *testing.T) {
	t.Parallel()

	repo := userRepoStub{
		getByEmailFn: func(_ context.Context, email string) (domain.User, error) {
			if email != "user@example.com" {
				t.Fatalf("GetByEmail() email = %s, want user@example.com", email)
			}
			return domain.User{}, domain.ErrUserNotFound
		},
	}

	svc := NewAuthService(auth.NewJWTManager("test-secret"), repo)

	_, err := svc.Login(context.Background(), " User@Example.com ", "secret-pass")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, domain.ErrInvalidCredentials)
	}
}

func TestAuthServiceLoginReturnsTokenForValidCredentials(t *testing.T) {
	t.Parallel()

	passwordHash, err := auth.HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	jwtManager := auth.NewJWTManager("test-secret")
	repo := userRepoStub{
		getByEmailFn: func(_ context.Context, email string) (domain.User, error) {
			if email != "user@example.com" {
				t.Fatalf("GetByEmail() email = %s, want user@example.com", email)
			}
			return domain.User{
				ID:           mustUUID("99999999-9999-9999-9999-999999999999"),
				Email:        email,
				Role:         domain.RoleUser,
				PasswordHash: passwordHash,
			}, nil
		},
	}

	svc := NewAuthService(jwtManager, repo)

	token, err := svc.Login(context.Background(), "User@example.com", "secret-pass")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if strings.TrimSpace(token) == "" {
		t.Fatal("Login() returned empty token")
	}

	claims, err := jwtManager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != "99999999-9999-9999-9999-999999999999" {
		t.Fatalf("claims.UserID = %s", claims.UserID)
	}
	if claims.Role != domain.RoleUser {
		t.Fatalf("claims.Role = %s, want %s", claims.Role, domain.RoleUser)
	}
}
