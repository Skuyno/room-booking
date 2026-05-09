package domain

import (
	"time"

	"github.com/google/uuid"
)

// Room — переговорка, корневая сущность бронирования. На неё ссылаются
// schedules.room_id и bookings.room_id. Description и Capacity — указатели,
// потому что в БД (и в API) это nullable-поля.
type Room struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Capacity    *int
	CreatedAt   time.Time
}
