// Package service содержит бизнес-логику приложения. Сервисы зависят от
// интерфейсов репозиториев, объявленных в этом файле, и не знают про
// конкретную реализацию (postgres). Это позволяет писать unit-тесты без
// БД через стабы и легко подменить реализацию хранилища.
//
// Композиция сервисов и репозиториев живёт в cmd/api/main.go.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

// UserRepository работает с таблицей users.
type UserRepository interface {
	// Create вставляет пользователя; возвращает domain.ErrEmailTaken при
	// нарушении уникальности email.
	Create(ctx context.Context, user domain.User) (domain.User, error)
	// GetByEmail находит пользователя по email; возвращает
	// domain.ErrUserNotFound, если такого нет.
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

// RoomRepository работает с таблицей rooms.
type RoomRepository interface {
	Create(ctx context.Context, room domain.Room) error
	List(ctx context.Context) ([]domain.Room, error)
	// Exists возвращает true, если комната с таким ID существует. Используется
	// сервисами для отдачи понятного ErrRoomNotFound вместо FK-ошибки от БД.
	Exists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

// ScheduleRepository работает с таблицей schedules. Расписание неизменяемо
// после создания и привязано к ровно одной комнате (unique room_id).
type ScheduleRepository interface {
	// Create вставляет расписание; возвращает domain.ErrScheduleExists при
	// нарушении уникальности room_id.
	Create(ctx context.Context, schedule domain.Schedule) error
	// GetByRoomID возвращает расписание комнаты или domain.ErrScheduleNotFound,
	// если у комнаты ещё нет расписания.
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, error)
}

// BookingRepository работает с таблицей bookings.
type BookingRepository interface {
	// Create вставляет бронь; возвращает domain.ErrSlotAlreadyBooked при
	// конфликте с уникальным partial-индексом
	// bookings_one_active_per_room_start_idx.
	Create(ctx context.Context, booking domain.Booking) error
	// GetByID возвращает бронь по ID или domain.ErrBookingNotFound.
	GetByID(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error)
	// Cancel помечает бронь cancelled и проставляет cancelled_at. Идемпотентна:
	// повторный вызов на уже отменённой брони не меняет cancelled_at.
	Cancel(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error)
	// ListByUserFuture возвращает только будущие брони пользователя
	// (start_at >= now), отсортированные по возрастанию времени начала.
	ListByUserFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error)
	// ListAll возвращает страницу всех броней (без фильтрации по пользователю
	// или статусу) и общее количество. Используется admin-ручкой /bookings/list.
	ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
	// ListActiveStartsByRoomBetween возвращает start_at активных броней комнаты
	// в полуоткрытом интервале [from, to). SlotService использует это, чтобы
	// убрать занятые окна из выдачи свободных слотов.
	ListActiveStartsByRoomBetween(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]time.Time, error)
}
