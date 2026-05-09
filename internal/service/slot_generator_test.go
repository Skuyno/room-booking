package service

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

func TestParseHHMM(t *testing.T) {
	t.Parallel()

	hour, minute, err := parseHHMM("09:30")
	if err != nil {
		t.Fatalf("parseHHMM() error = %v", err)
	}
	if hour != 9 || minute != 30 {
		t.Fatalf("parseHHMM() = (%d, %d), want (9, 30)", hour, minute)
	}

	_, _, err = parseHHMM("25:00")
	if err == nil {
		t.Fatal("parseHHMM() expected error for invalid time")
	}
}

func TestNormalizeAndValidateDays(t *testing.T) {
	t.Parallel()

	days, err := NormalizeAndValidateDays([]int16{5, 1, 5, 3})
	if err != nil {
		t.Fatalf("NormalizeAndValidateDays() error = %v", err)
	}
	if !reflect.DeepEqual(days, []int16{1, 3, 5}) {
		t.Fatalf("days = %v, want [1 3 5]", days)
	}

	_, err = NormalizeAndValidateDays([]int16{0})
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("NormalizeAndValidateDays() error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestValidateScheduleWindow(t *testing.T) {
	t.Parallel()

	if err := validateScheduleWindow("09:00", "10:30"); err != nil {
		t.Fatalf("validateScheduleWindow() error = %v", err)
	}

	if err := validateScheduleWindow("09:15", "10:30"); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("validateScheduleWindow() error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestGenerateSlotsForDate(t *testing.T) {
	t.Parallel()

	// 6 апреля 2026 — понедельник.
	monday := time.Date(2026, 4, 6, 15, 0, 0, 0, time.FixedZone("UTC+4", 4*60*60))
	schedule := domain.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int16{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	slots, err := GenerateSlotsForDate(schedule, monday)
	if err != nil {
		t.Fatalf("GenerateSlotsForDate() error = %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("len(slots) = %d, want 2", len(slots))
	}
	if !slots[0].StartAt.Equal(time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("slots[0].StartAt = %v", slots[0].StartAt)
	}
	if !slots[1].EndAt.Equal(time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("slots[1].EndAt = %v", slots[1].EndAt)
	}
}

func TestGenerateSlotsForDateReturnsEmptyForOffDay(t *testing.T) {
	t.Parallel()

	tuesday := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)
	schedule := domain.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int16{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	slots, err := GenerateSlotsForDate(schedule, tuesday)
	if err != nil {
		t.Fatalf("GenerateSlotsForDate() error = %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("len(slots) = %d, want 0", len(slots))
	}
}

func TestIsStartAtWithinSchedule(t *testing.T) {
	t.Parallel()

	schedule := domain.Schedule{
		DaysOfWeek: []int16{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	monday := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)

	if !IsStartAtWithinSchedule(schedule, monday) {
		t.Fatal("Monday 09:00 must be within schedule")
	}
	if IsStartAtWithinSchedule(schedule, monday.Add(15*time.Minute)) {
		t.Fatal("Monday 09:15 is not aligned to 30 minutes")
	}
	if IsStartAtWithinSchedule(schedule, monday.Add(time.Hour)) {
		t.Fatal("Monday 10:00 has no full slot before window end")
	}
	if IsStartAtWithinSchedule(schedule, monday.AddDate(0, 0, 1)) {
		t.Fatal("Tuesday 09:00 is outside daysOfWeek")
	}
}
