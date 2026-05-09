package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

// RoomRepository — pgx-реализация service.RoomRepository.
type RoomRepository struct {
	q DBTX
}

// NewRoomRepository создаёт репозиторий поверх любого DBTX (пул или транзакция).
func NewRoomRepository(q DBTX) *RoomRepository {
	return &RoomRepository{q: q}
}

// Create вставляет переговорку. ID, Description и Capacity берутся из
// переданной структуры; created_at проставляется DEFAULT-ом БД.
func (r *RoomRepository) Create(ctx context.Context, room domain.Room) error {
	_, err := r.q.Exec(ctx, `
		insert into rooms (id, name, description, capacity)
		values ($1, $2, $3, $4)
	`, room.ID, room.Name, room.Description, room.Capacity)
	if err != nil {
		return fmt.Errorf("insert room: %w", err)
	}
	return nil
}

// Exists возвращает true, если переговорка с таким ID есть в БД. Используется
// сервисами для отдачи ErrRoomNotFound вместо непрозрачной FK-ошибки.
func (r *RoomRepository) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	var exists bool
	err := r.q.QueryRow(ctx, `select exists(select 1 from rooms where id = $1)`, roomID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check room exists: %w", err)
	}
	return exists, nil
}

// List возвращает все переговорки в порядке возрастания created_at.
// Description и Capacity скан-ятся через sql.NullString/NullInt32, чтобы
// корректно обработать NULL.
func (r *RoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	rows, err := r.q.Query(ctx, `
		select id, name, description, capacity, created_at
		from rooms
		order by created_at asc
	`)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var result []domain.Room
	for rows.Next() {
		var room domain.Room
		var description sql.NullString
		var capacity sql.NullInt32

		if err := rows.Scan(
			&room.ID,
			&room.Name,
			&description,
			&capacity,
			&room.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		if description.Valid {
			room.Description = &description.String
		}
		if capacity.Valid {
			v := int(capacity.Int32)
			room.Capacity = &v
		}

		result = append(result, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return result, nil
}
