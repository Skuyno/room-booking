package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Skuyno/room-booking/internal/domain"
)

// ScheduleRepository — pgx-реализация service.ScheduleRepository.
// Расписаний в системе ровно по одному на комнату (unique room_id).
type ScheduleRepository struct {
	q DBTX
}

// NewScheduleRepository создаёт репозиторий поверх любого DBTX.
func NewScheduleRepository(q DBTX) *ScheduleRepository {
	return &ScheduleRepository{q: q}
}

// Create вставляет расписание. При нарушении уникальности room_id
// (constraint schedules_room_unique) возвращает domain.ErrScheduleExists.
//
// Имя констрейнта матчится по подстроке в тексте ошибки — это хрупкая, но
// работающая практика, потому что pgx-ошибки нарушений такого рода не
// разделяют по типу. Если когда-нибудь имя констрейнта поменяется,
// надо обновить и эту проверку.
func (r *ScheduleRepository) Create(ctx context.Context, schedule domain.Schedule) error {
	_, err := r.q.Exec(ctx, `
		insert into schedules (id, room_id, days_of_week, start_time, end_time)
		values ($1, $2, $3, $4, $5)
	`, schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime)
	if err != nil {
		if strings.Contains(err.Error(), "schedules_room_unique") {
			return domain.ErrScheduleExists
		}
		return fmt.Errorf("insert schedule: %w", err)
	}
	return nil
}

// GetByRoomID возвращает расписание комнаты или domain.ErrScheduleNotFound,
// если у комнаты ещё нет расписания.
//
// Время хранится в Postgres как time without time zone и сериализуется
// как "HH:MM:SS"; кастуем в text и обрезаем секунды через trimSeconds,
// чтобы домен оперировал привычным "HH:MM".
func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, error) {
	var schedule domain.Schedule

	err := r.q.QueryRow(ctx, `
		select id, room_id, days_of_week, start_time::text, end_time::text, created_at
		from schedules
		where room_id = $1
	`, roomID).Scan(
		&schedule.ID,
		&schedule.RoomID,
		&schedule.DaysOfWeek,
		&schedule.StartTime,
		&schedule.EndTime,
		&schedule.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Schedule{}, domain.ErrScheduleNotFound
		}
		return domain.Schedule{}, fmt.Errorf("get schedule by room id: %w", err)
	}

	schedule.StartTime = trimSeconds(schedule.StartTime)
	schedule.EndTime = trimSeconds(schedule.EndTime)

	return schedule, nil
}

// trimSeconds приводит time-значение PostgreSQL вида "09:00:00" к "09:00".
// Если строка короче 5 символов или нет двух двоеточий — возвращается как есть.
func trimSeconds(value string) string {
	if len(value) >= 5 && strings.Count(value, ":") >= 2 {
		return value[:5]
	}
	return value
}
