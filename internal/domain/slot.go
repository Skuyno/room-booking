package domain

import (
	"time"

	"github.com/google/uuid"
)

// Slot — это виртуальное окно бронирования: оно не хранится в БД, а
// вычисляется из расписания комнаты на конкретную дату функцией
// GenerateSlotsForDate. Поэтому у него нет ID и CreatedAt — два слота с
// одинаковыми (RoomID, StartAt) идентичны.
type Slot struct {
	RoomID  uuid.UUID
	StartAt time.Time
	EndAt   time.Time
}
