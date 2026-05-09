package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

func TestSlotServiceListAvailableByRoomAndDateReturnsRoomNotFound(t *testing.T) {
	t.Parallel()

	svc := NewSlotService(
		roomRepoStub{
			existsFn: func(_ context.Context, roomID uuid.UUID) (bool, error) {
				return false, nil
			},
		},
		scheduleRepoStub{},
		bookingRepoStub{},
	)

	_, err := svc.ListAvailableByRoomAndDate(context.Background(), uuid.New(), time.Now())
	if !errors.Is(err, domain.ErrRoomNotFound) {
		t.Fatalf("ListAvailableByRoomAndDate() error = %v, want %v", err, domain.ErrRoomNotFound)
	}
}

func TestSlotServiceListAvailableByRoomAndDateReturnsEmptyWhenNoSchedule(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()

	svc := NewSlotService(
		roomRepoStub{
			existsFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return true, nil
			},
		},
		scheduleRepoStub{
			getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, error) {
				return domain.Schedule{}, domain.ErrScheduleNotFound
			},
		},
		bookingRepoStub{},
	)

	slots, err := svc.ListAvailableByRoomAndDate(context.Background(), roomID, time.Now())
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate() error = %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("len(slots) = %d, want 0", len(slots))
	}
}

func TestSlotServiceListAvailableByRoomAndDateFiltersBookedSlots(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	date := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC) // Monday
	bookedStart := time.Date(2026, 4, 6, 9, 30, 0, 0, time.UTC)

	svc := NewSlotService(
		roomRepoStub{
			existsFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
				return true, nil
			},
		},
		scheduleRepoStub{
			getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, error) {
				return domain.Schedule{
					ID:         uuid.New(),
					RoomID:     roomID,
					DaysOfWeek: []int16{1},
					StartTime:  "09:00",
					EndTime:    "10:30",
				}, nil
			},
		},
		bookingRepoStub{
			listActiveStartsByRoomBetweenFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]time.Time, error) {
				return []time.Time{bookedStart}, nil
			},
		},
	)

	slots, err := svc.ListAvailableByRoomAndDate(context.Background(), roomID, date)
	if err != nil {
		t.Fatalf("ListAvailableByRoomAndDate() error = %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("len(slots) = %d, want 2", len(slots))
	}
	for _, slot := range slots {
		if slot.StartAt.Equal(bookedStart) {
			t.Fatalf("booked slot %v should be filtered out", slot.StartAt)
		}
	}
}
