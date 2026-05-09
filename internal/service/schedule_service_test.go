package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

func TestScheduleServiceCreateRejectsInvalidWindow(t *testing.T) {
	t.Parallel()

	svc := NewScheduleService(roomRepoStub{}, scheduleRepoStub{})
	roomID := uuid.New()

	_, err := svc.Create(context.Background(), roomID, []int16{1}, "09:15", "10:00")
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("Create() invalid start time error = %v, want %v", err, domain.ErrInvalidRequest)
	}

	_, err = svc.Create(context.Background(), roomID, []int16{1}, "10:00", "09:30")
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("Create() invalid range error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestScheduleServiceCreateReturnsRoomNotFound(t *testing.T) {
	t.Parallel()

	svc := NewScheduleService(
		roomRepoStub{
			existsFn: func(_ context.Context, roomID uuid.UUID) (bool, error) {
				return false, nil
			},
		},
		scheduleRepoStub{},
	)

	_, err := svc.Create(context.Background(), uuid.New(), []int16{1}, "09:00", "10:00")
	if !errors.Is(err, domain.ErrRoomNotFound) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrRoomNotFound)
	}
}

func TestScheduleServiceCreatePersistsAndNormalizesDays(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	var capturedSchedule domain.Schedule

	svc := NewScheduleService(
		roomRepoStub{
			existsFn: func(_ context.Context, gotRoomID uuid.UUID) (bool, error) {
				if gotRoomID != roomID {
					t.Fatalf("Exists() roomID = %s, want %s", gotRoomID, roomID)
				}
				return true, nil
			},
		},
		scheduleRepoStub{
			createFn: func(_ context.Context, schedule domain.Schedule) error {
				capturedSchedule = schedule
				return nil
			},
		},
	)

	schedule, err := svc.Create(context.Background(), roomID, []int16{3, 1, 1, 5}, "09:00", "10:00")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !reflect.DeepEqual(capturedSchedule.DaysOfWeek, []int16{1, 3, 5}) {
		t.Fatalf("capturedSchedule.DaysOfWeek = %v, want [1 3 5]", capturedSchedule.DaysOfWeek)
	}
	if schedule.RoomID != roomID {
		t.Fatalf("returned schedule roomID = %s, want %s", schedule.RoomID, roomID)
	}
	if schedule.ID == uuid.Nil {
		t.Fatal("schedule.ID must be generated")
	}
}
