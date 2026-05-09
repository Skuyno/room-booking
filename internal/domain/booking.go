package domain

import (
	"time"

	"github.com/google/uuid"
)

// Строковые коды статусов брони. Используются в API и в логике сервисов.
const (
	BookingStatusActive    = "active"
	BookingStatusCancelled = "cancelled"
)

// Числовые ID статусов брони. Хранятся в bookings.status_id и в
// таблице-справочнике booking_statuses. Активная бронь занимает партишн в
// уникальном индексе bookings_one_active_per_room_start_idx.
const (
	BookingStatusIDActive    int16 = 1
	BookingStatusIDCancelled int16 = 2
)

// Booking — бронь конкретного 30-минутного окна в конкретной комнате.
// Привязка к расписанию подразумевается через RoomID + StartAt: проверка
// валидности времени делается в BookingService через
// IsStartAtWithinSchedule. ConferenceLink и CancelledAt опциональны.
type Booking struct {
	ID             uuid.UUID
	RoomID         uuid.UUID
	UserID         uuid.UUID
	StatusID       int16
	Status         string
	StartAt        time.Time
	EndAt          time.Time
	ConferenceLink *string
	CreatedAt      time.Time
	CancelledAt    *time.Time
}

// BookingStatusCodeByID конвертирует числовой ID статуса в строковой код.
// Возвращает пустую строку для неизвестного ID.
func BookingStatusCodeByID(id int16) string {
	switch id {
	case BookingStatusIDActive:
		return BookingStatusActive
	case BookingStatusIDCancelled:
		return BookingStatusCancelled
	default:
		return ""
	}
}

// BookingStatusIDByCode конвертирует строковой код статуса в числовой ID.
// Возвращает 0 для неизвестного кода.
func BookingStatusIDByCode(code string) int16 {
	switch code {
	case BookingStatusActive:
		return BookingStatusIDActive
	case BookingStatusCancelled:
		return BookingStatusIDCancelled
	default:
		return 0
	}
}
