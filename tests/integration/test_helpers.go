package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Skuyno/room-booking/internal/domain"
	"github.com/Skuyno/room-booking/internal/storage/postgres"
)

var (
	adminID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	if dsn == "" {
		t.Skip("TEST_DB_DSN or DB_DSN is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	if err := postgres.RunMigrations(context.Background(), pool); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return pool
}

func resetDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		truncate table bookings, schedules, rooms, users, booking_statuses, user_roles
		restart identity cascade;
	`)
	if err != nil {
		t.Fatalf("truncate db: %v", err)
	}

	_, err = pool.Exec(context.Background(), `
		insert into user_roles (id, code)
		values
			($1, 'admin'),
			($2, 'user')
	`, domain.RoleIDAdmin, domain.RoleIDUser)
	if err != nil {
		t.Fatalf("seed roles: %v", err)
	}

	_, err = pool.Exec(context.Background(), `
		insert into booking_statuses (id, code)
		values
			($1, 'active'),
			($2, 'cancelled')
	`, domain.BookingStatusIDActive, domain.BookingStatusIDCancelled)
	if err != nil {
		t.Fatalf("seed booking statuses: %v", err)
	}

	_, err = pool.Exec(context.Background(), `
		insert into users (id, email, role_id)
		values
			($1, 'admin@example.com', $3),
			($2, 'user@example.com', $4)
	`, adminID, userID, domain.RoleIDAdmin, domain.RoleIDUser)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}
}

func isoDayFromTime(ts time.Time) int16 {
	switch ts.UTC().Weekday() {
	case time.Monday:
		return 1
	case time.Tuesday:
		return 2
	case time.Wednesday:
		return 3
	case time.Thursday:
		return 4
	case time.Friday:
		return 5
	case time.Saturday:
		return 6
	case time.Sunday:
		return 7
	default:
		return 0
	}
}
