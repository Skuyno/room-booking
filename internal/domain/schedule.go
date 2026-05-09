package domain

import (
	"time"

	"github.com/google/uuid"
)

// Schedule — правило доступности комнаты: какие дни недели и в какое окно.
// На комнату приходится максимум одно расписание (уникальный constraint
// schedules_room_unique). После создания не редактируется.
//
// DaysOfWeek хранит ISO-номера дней (1=Mon ... 7=Sun). StartTime/EndTime —
// строки HH:MM в UTC, минуты обязательно кратны 30 (валидируется в сервисе
// и в check-constraint БД). Слоты по этому расписанию вычисляются на лету
// функцией GenerateSlotsForDate из пакета service.
type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek []int16
	StartTime  string
	EndTime    string
	CreatedAt  time.Time
}
