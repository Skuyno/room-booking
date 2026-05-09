package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Skuyno/room-booking/internal/domain"
	"github.com/Skuyno/room-booking/internal/service"
	"github.com/Skuyno/room-booking/internal/storage/postgres"
)

func TestCancelBookingFlow(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	defer pool.Close()

	resetDB(t, pool)

	roomRepo := postgres.NewRoomRepository(pool)
	bookingRepo := postgres.NewBookingRepository(pool)
	scheduleRepo := postgres.NewScheduleRepository(pool)

	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo)
	slotService := service.NewSlotService(roomRepo, scheduleRepo, bookingRepo)
	bookingService := service.NewBookingService(scheduleRepo, bookingRepo)

	room, err := roomService.Create(ctx, "Room B", nil, nil)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	targetDay := time.Now().UTC().AddDate(0, 0, 1)
	day := isoDayFromTime(targetDay)

	_, err = scheduleService.Create(ctx, room.ID, []int16{day}, "09:00", "10:00")
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	slots, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, targetDay)
	if err != nil {
		t.Fatalf("list slots: %v", err)
	}
	if len(slots) == 0 {
		t.Fatal("expected slots")
	}

	booking, err := bookingService.Create(ctx, userID, domain.RoleUser, room.ID, slots[0].StartAt, false)
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	cancelled, err := bookingService.Cancel(ctx, userID, domain.RoleUser, booking.ID)
	if err != nil {
		t.Fatalf("cancel booking: %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled booking, got %s", cancelled.Status)
	}

	cancelledAgain, err := bookingService.Cancel(ctx, userID, domain.RoleUser, booking.ID)
	if err != nil {
		t.Fatalf("cancel booking second time: %v", err)
	}
	if cancelledAgain.Status != domain.BookingStatusCancelled {
		t.Fatalf("expected cancelled booking second time, got %s", cancelledAgain.Status)
	}

	availableAfterCancel, err := slotService.ListAvailableByRoomAndDate(ctx, room.ID, targetDay)
	if err != nil {
		t.Fatalf("list slots after cancel: %v", err)
	}
	if len(availableAfterCancel) != 2 {
		t.Fatalf("expected 2 available slots after cancel, got %d", len(availableAfterCancel))
	}
}
