package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Skuyno/room-booking/internal/domain"
)

// UserRepository — pgx-реализация service.UserRepository.
type UserRepository struct {
	q DBTX
}

// NewUserRepository создаёт репозиторий поверх любого DBTX.
func NewUserRepository(q DBTX) *UserRepository {
	return &UserRepository{q: q}
}

// Create вставляет пользователя и возвращает запись с фактическими
// timestamp-ами и ролью. При нарушении уникальности email
// (PostgreSQL код "23505") возвращает domain.ErrEmailTaken.
func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	var created domain.User

	err := r.q.QueryRow(ctx, `
		insert into users (id, email, role_id, password_hash)
		values ($1, $2, $3, $4)
		returning id, email, role_id, password_hash, created_at
	`, user.ID, user.Email, user.RoleID, user.PasswordHash).Scan(
		&created.ID,
		&created.Email,
		&created.RoleID,
		&created.PasswordHash,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrEmailTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	created.Role = domain.RoleCodeByID(created.RoleID)
	return created, nil
}

// GetByEmail находит пользователя по email или возвращает
// domain.ErrUserNotFound. PasswordHash может быть NULL для seed-юзеров под
// /dummyLogin — тогда поле остаётся пустой строкой.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	var passwordHash sql.NullString

	err := r.q.QueryRow(ctx, `
		select id, email, role_id, password_hash, created_at
		from users
		where email = $1
	`, email).Scan(
		&user.ID,
		&user.Email,
		&user.RoleID,
		&passwordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	user.Role = domain.RoleCodeByID(user.RoleID)

	return user, nil
}

// GetByID находит пользователя по UUID или возвращает domain.ErrUserNotFound.
func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	var user domain.User
	var passwordHash sql.NullString

	err := r.q.QueryRow(ctx, `
		select id, email, role_id, password_hash, created_at
		from users
		where id = $1
	`, userID).Scan(
		&user.ID,
		&user.Email,
		&user.RoleID,
		&passwordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	user.Role = domain.RoleCodeByID(user.RoleID)

	return user, nil
}
