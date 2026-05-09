package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

func nextMonday(t time.Time) time.Time {
	t = t.UTC()
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

func mondayAt(hour, minute int) time.Time {
	day := nextMonday(time.Now().UTC().AddDate(0, 0, 1))
	return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, time.UTC)
}

func openSchedule(roomID uuid.UUID) domain.Schedule {
	return domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int16{1, 2, 3, 4, 5, 6, 7},
		StartTime:  "00:00",
		EndTime:    "23:30",
	}
}

func TestBookingServiceCreateRejectsNonUserRole(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(scheduleRepoStub{}, bookingRepoStub{})

	_, err := svc.Create(context.Background(), uuid.New(), domain.RoleAdmin, uuid.New(), time.Now().Add(time.Hour), false)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrForbidden)
	}
}

func TestBookingServiceCreateRejectsPastSlot(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(scheduleRepoStub{}, bookingRepoStub{})

	_, err := svc.Create(
		context.Background(),
		uuid.New(),
		domain.RoleUser,
		uuid.New(),
		time.Now().UTC().Add(-time.Hour),
		false,
	)
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestBookingServiceCreateReturnsSlotNotFoundWhenNoSchedule(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(
		scheduleRepoStub{
			getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, error) {
				return domain.Schedule{}, domain.ErrScheduleNotFound
			},
		},
		bookingRepoStub{},
	)

	_, err := svc.Create(
		context.Background(),
		uuid.New(),
		domain.RoleUser,
		uuid.New(),
		mondayAt(10, 0),
		false,
	)
	if !errors.Is(err, domain.ErrSlotNotFound) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrSlotNotFound)
	}
}

func TestBookingServiceCreateRejectsStartAtOutsideSchedule(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	svc := NewBookingService(
		scheduleRepoStub{
			getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, error) {
				return domain.Schedule{
					ID:         uuid.New(),
					RoomID:     roomID,
					DaysOfWeek: []int16{1},
					StartTime:  "09:00",
					EndTime:    "10:00",
				}, nil
			},
		},
		bookingRepoStub{},
	)

	// 12:00 is outside the 09:00–10:00 window
	startAt := mondayAt(12, 0)

	_, err := svc.Create(context.Background(), uuid.New(), domain.RoleUser, roomID, startAt, false)
	if !errors.Is(err, domain.ErrSlotNotFound) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrSlotNotFound)
	}
}

func TestBookingServiceCreatePersistsActiveBooking(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	roomID := uuid.New()
	startAt := mondayAt(10, 0)
	var created domain.Booking

	svc := NewBookingService(
		scheduleRepoStub{
			getByRoomIDFn: func(_ context.Context, gotRoomID uuid.UUID) (domain.Schedule, error) {
				if gotRoomID != roomID {
					t.Fatalf("GetByRoomID() roomID = %s, want %s", gotRoomID, roomID)
				}
				return openSchedule(roomID), nil
			},
		},
		bookingRepoStub{
			createFn: func(_ context.Context, booking domain.Booking) error {
				created = booking
				return nil
			},
		},
	)

	booking, err := svc.Create(context.Background(), userID, domain.RoleUser, roomID, startAt, true)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.UserID != userID || created.RoomID != roomID {
		t.Fatalf("created booking = %+v", created)
	}
	if !created.StartAt.Equal(startAt) {
		t.Fatalf("created.StartAt = %v, want %v", created.StartAt, startAt)
	}
	if !created.EndAt.Equal(startAt.Add(SlotDuration)) {
		t.Fatalf("created.EndAt = %v, want %v", created.EndAt, startAt.Add(SlotDuration))
	}
	if created.Status != domain.BookingStatusActive {
		t.Fatalf("created.Status = %s, want %s", created.Status, domain.BookingStatusActive)
	}
	if created.ConferenceLink == nil || !strings.HasPrefix(*created.ConferenceLink, "https://conference.mock/") {
		t.Fatalf("created.ConferenceLink = %v", created.ConferenceLink)
	}
	if booking.ID == uuid.Nil {
		t.Fatal("booking.ID must be generated")
	}
}

func TestBookingServiceCancelRejectsAnotherUsersBooking(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	callerID := uuid.New()
	cancelCalled := false

	svc := NewBookingService(
		scheduleRepoStub{},
		bookingRepoStub{
			getByIDFn: func(_ context.Context, bookingID uuid.UUID) (domain.Booking, error) {
				return domain.Booking{ID: bookingID, UserID: ownerID}, nil
			},
			cancelFn: func(_ context.Context, bookingID uuid.UUID) (domain.Booking, error) {
				cancelCalled = true
				return domain.Booking{}, nil
			},
		},
	)

	_, err := svc.Cancel(context.Background(), callerID, domain.RoleUser, uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Cancel() error = %v, want %v", err, domain.ErrForbidden)
	}
	if cancelCalled {
		t.Fatal("Cancel() repository method must not be called for another user's booking")
	}
}

func TestBookingServiceListAllNormalizesPagination(t *testing.T) {
	t.Parallel()

	var gotPage int
	var gotPageSize int

	svc := NewBookingService(
		scheduleRepoStub{},
		bookingRepoStub{
			listAllFn: func(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
				gotPage = page
				gotPageSize = pageSize
				return nil, 0, nil
			},
		},
	)

	_, _, err := svc.ListAll(context.Background(), domain.RoleAdmin, 0, 0)
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if gotPage != 1 || gotPageSize != 20 {
		t.Fatalf("pagination = (%d, %d), want (1, 20)", gotPage, gotPageSize)
	}
}

func TestBookingServiceListAllRejectsTooLargePageSize(t *testing.T) {
	t.Parallel()

	svc := NewBookingService(scheduleRepoStub{}, bookingRepoStub{})

	_, _, err := svc.ListAll(context.Background(), domain.RoleAdmin, 1, 101)
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("ListAll() error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}
