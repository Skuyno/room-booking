package http

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

func TestRoomToAPIIncludesOptionalFields(t *testing.T) {
	t.Parallel()

	description := "Main room"
	capacity := 8
	createdAt := time.Date(2026, 4, 5, 10, 0, 0, 0, time.FixedZone("UTC+4", 4*60*60))

	room := roomToAPI(domain.Room{
		ID:          uuid.New(),
		Name:        "Omega",
		Description: &description,
		Capacity:    &capacity,
		CreatedAt:   createdAt,
	})

	if room.Description == nil || *room.Description != description {
		t.Fatalf("Description = %v, want %s", room.Description, description)
	}
	if room.Capacity == nil || *room.Capacity != capacity {
		t.Fatalf("Capacity = %v, want %d", room.Capacity, capacity)
	}
	if room.CreatedAt == nil || room.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt = %v, want UTC time", room.CreatedAt)
	}
}

func TestScheduleAndSlotMappingUseUTC(t *testing.T) {
	t.Parallel()

	scheduleID := uuid.New()
	roomID := uuid.New()
	slot := slotToAPI(domain.Slot{
		RoomID:  roomID,
		StartAt: time.Date(2026, 4, 5, 12, 0, 0, 0, time.FixedZone("UTC+4", 4*60*60)),
		EndAt:   time.Date(2026, 4, 5, 12, 30, 0, 0, time.FixedZone("UTC+4", 4*60*60)),
	})
	schedule := scheduleToAPI(domain.Schedule{
		ID:         scheduleID,
		RoomID:     roomID,
		DaysOfWeek: []int16{1, 3, 5},
		StartTime:  "09:00",
		EndTime:    "18:00",
	})

	if schedule.Id == nil || *schedule.Id != scheduleID {
		t.Fatalf("schedule.Id = %v, want %s", schedule.Id, scheduleID)
	}
	if len(schedule.DaysOfWeek) != 3 || schedule.DaysOfWeek[1] != 3 {
		t.Fatalf("schedule.DaysOfWeek = %v", schedule.DaysOfWeek)
	}
	if slot.Start.Location() != time.UTC || slot.End.Location() != time.UTC {
		t.Fatalf("slot times must be UTC: start=%v end=%v", slot.Start, slot.End)
	}
}

func TestBookingToAPIOmitsNilConferenceLink(t *testing.T) {
	t.Parallel()

	booking := bookingToAPI(domain.Booking{
		ID:        uuid.New(),
		RoomID:    uuid.New(),
		UserID:    uuid.New(),
		Status:    domain.BookingStatusActive,
		StartAt:   time.Now().UTC(),
		EndAt:     time.Now().UTC().Add(30 * time.Minute),
		CreatedAt: time.Now(),
	})

	if booking.ConferenceLink != nil {
		t.Fatalf("ConferenceLink = %v, want nil", booking.ConferenceLink)
	}
	if booking.CreatedAt == nil {
		t.Fatal("CreatedAt must be set")
	}
}

func TestTimePtrReturnsNilForZeroTime(t *testing.T) {
	t.Parallel()

	if timePtr(time.Time{}) != nil {
		t.Fatal("timePtr(zero) must return nil")
	}
	if uuidPtr(uuid.Nil) == nil {
		t.Fatal("uuidPtr() must always return pointer")
	}
}
