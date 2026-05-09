package http

import (
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/domain"
)

// userToAPI конвертирует domain.User в api.User. PasswordHash и RoleID
// наружу не отдаются — только id, email, role и опционально createdAt.
func userToAPI(user domain.User) api.User {
	return api.User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		Role:      api.UserRole(user.Role),
		CreatedAt: timePtr(user.CreatedAt),
	}
}

// roomToAPI конвертирует domain.Room в api.Room, сохраняя nullable-поля
// (Description, Capacity) как nil-указатели если они не заданы.
func roomToAPI(room domain.Room) api.Room {
	resp := api.Room{
		Id:   room.ID,
		Name: room.Name,
	}

	if room.Description != nil {
		resp.Description = room.Description
	}
	if room.Capacity != nil {
		resp.Capacity = room.Capacity
	}
	resp.CreatedAt = timePtr(room.CreatedAt)

	return resp
}

// roomsToAPI batch-конвертирует слайс комнат.
func roomsToAPI(items []domain.Room) []api.Room {
	result := make([]api.Room, 0, len(items))
	for _, item := range items {
		result = append(result, roomToAPI(item))
	}
	return result
}

// scheduleToAPI конвертирует domain.Schedule в api.Schedule.
// DaysOfWeek приходится копировать в []int, потому что доменное
// представление — []int16 (под smallint в БД).
func scheduleToAPI(schedule domain.Schedule) api.Schedule {
	days := make([]int, 0, len(schedule.DaysOfWeek))
	for _, d := range schedule.DaysOfWeek {
		days = append(days, int(d))
	}

	return api.Schedule{
		Id:         uuidPtr(schedule.ID),
		RoomId:     schedule.RoomID,
		DaysOfWeek: days,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
	}
}

// slotToAPI конвертирует domain.Slot. ID у слота нет — это виртуальная
// сущность, идентифицируемая парой (roomId, start).
func slotToAPI(slot domain.Slot) api.Slot {
	return api.Slot{
		RoomId: slot.RoomID,
		Start:  slot.StartAt.UTC(),
		End:    slot.EndAt.UTC(),
	}
}

// slotsToAPI batch-конвертирует слайс слотов.
func slotsToAPI(items []domain.Slot) []api.Slot {
	result := make([]api.Slot, 0, len(items))
	for _, item := range items {
		result = append(result, slotToAPI(item))
	}
	return result
}

// bookingToAPI конвертирует domain.Booking. StartAt и EndAt всегда
// отдаются в UTC, чтобы клиенту не приходилось гадать про таймзону.
func bookingToAPI(booking domain.Booking) api.Booking {
	resp := api.Booking{
		Id:      booking.ID,
		RoomId:  booking.RoomID,
		UserId:  booking.UserID,
		Status:  api.BookingStatus(booking.Status),
		StartAt: booking.StartAt.UTC(),
		EndAt:   booking.EndAt.UTC(),
	}

	if booking.ConferenceLink != nil {
		resp.ConferenceLink = booking.ConferenceLink
	}
	resp.CreatedAt = timePtr(booking.CreatedAt)

	return resp
}

// bookingsToAPI batch-конвертирует слайс броней.
func bookingsToAPI(items []domain.Booking) []api.Booking {
	result := make([]api.Booking, 0, len(items))
	for _, item := range items {
		result = append(result, bookingToAPI(item))
	}
	return result
}

// timePtr возвращает указатель на UTC-копию t. Для нулевого времени
// возвращает nil — это нужно, чтобы в JSON отдать null в полях типа
// createdAt, когда сервис не заполнил время (например, при mock-данных).
func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	v := t.UTC()
	return &v
}

// uuidPtr оборачивает uuid в указатель. Использует копию, чтобы вызывающий
// не мог через указатель модифицировать переданный id (это бы повлияло на
// исходную доменную сущность).
func uuidPtr(id uuid.UUID) *uuid.UUID {
	return &id
}
