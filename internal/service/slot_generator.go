package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Skuyno/room-booking/internal/domain"
)

// SlotDuration — фиксированная длительность одного слота бронирования.
// Это же значение зашито в check-constraint bookings_duration_30m в БД,
// поэтому менять его нужно одновременно в коде и в миграции.
const SlotDuration = 30 * time.Minute

// GenerateSlotsForDate возвращает все возможные слоты заданного расписания
// на конкретную дату. Если день недели не входит в DaysOfWeek — возвращается
// nil. Все StartAt/EndAt возвращаются в UTC.
//
// Функция ничего не знает про существующие брони — фильтрация занятых
// слотов делается выше, в SlotService.
func GenerateSlotsForDate(schedule domain.Schedule, date time.Time) ([]domain.Slot, error) {
	startHour, startMinute, err := parseHHMM(schedule.StartTime)
	if err != nil {
		return nil, fmt.Errorf("parse start time: %w", err)
	}

	endHour, endMinute, err := parseHHMM(schedule.EndTime)
	if err != nil {
		return nil, fmt.Errorf("parse end time: %w", err)
	}

	dayUTC := truncateToDayUTC(date)

	allowedDays := normalizeDays(schedule.DaysOfWeek)
	isoDay := weekdayToISO(dayUTC.Weekday())
	if !allowedDays[isoDay] {
		return nil, nil
	}

	windowStart := time.Date(dayUTC.Year(), dayUTC.Month(), dayUTC.Day(), startHour, startMinute, 0, 0, time.UTC)
	windowEnd := time.Date(dayUTC.Year(), dayUTC.Month(), dayUTC.Day(), endHour, endMinute, 0, 0, time.UTC)

	var slots []domain.Slot
	for cursor := windowStart; cursor.Before(windowEnd); cursor = cursor.Add(SlotDuration) {
		slotEnd := cursor.Add(SlotDuration)
		if slotEnd.After(windowEnd) {
			break
		}
		slots = append(slots, domain.Slot{
			RoomID:  schedule.RoomID,
			StartAt: cursor,
			EndAt:   slotEnd,
		})
	}

	return slots, nil
}

// IsStartAtWithinSchedule проверяет, что startAt валиден относительно
// расписания и подходит для бронирования. Используется BookingService для
// валидации запроса POST /bookings/create.
//
// Условия валидности:
//   - startAt в UTC, секунды и наносекунды нулевые, минуты кратны 30;
//   - день недели входит в schedule.DaysOfWeek;
//   - окно [startAt, startAt+SlotDuration] целиком помещается в
//     [schedule.StartTime, schedule.EndTime].
//
// Возвращает true только если все условия выполнены.
func IsStartAtWithinSchedule(schedule domain.Schedule, startAt time.Time) bool {
	startAt = startAt.UTC()

	if startAt.Second() != 0 || startAt.Nanosecond() != 0 {
		return false
	}
	if startAt.Minute()%30 != 0 {
		return false
	}

	allowedDays := normalizeDays(schedule.DaysOfWeek)
	if !allowedDays[weekdayToISO(startAt.Weekday())] {
		return false
	}

	startHour, startMinute, err := parseHHMM(schedule.StartTime)
	if err != nil {
		return false
	}
	endHour, endMinute, err := parseHHMM(schedule.EndTime)
	if err != nil {
		return false
	}

	currentMinutes := startAt.Hour()*60 + startAt.Minute()
	windowStartMinutes := startHour*60 + startMinute
	windowEndMinutes := endHour*60 + endMinute

	if currentMinutes < windowStartMinutes {
		return false
	}
	if currentMinutes+int(SlotDuration/time.Minute) > windowEndMinutes {
		return false
	}

	return true
}

// truncateToDayUTC переводит t в UTC и возвращает начало календарных суток.
// Нужно, чтобы при формировании окон слотов на дату не зависеть от
// таймзоны клиента.
func truncateToDayUTC(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// parseHHMM парсит строку формата "HH:MM" в часы и минуты. Возвращает ошибку
// при невалидном формате или часах/минутах вне диапазона.
func parseHHMM(value string) (int, int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse hour: %w", err)
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse minute: %w", err)
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("time is out of range")
	}

	return hour, minute, nil
}

// weekdayToISO конвертирует time.Weekday (Sunday=0..Saturday=6) в ISO-8601
// номер дня недели (Monday=1..Sunday=7), который используется в schedules.days_of_week.
func weekdayToISO(wd time.Weekday) int16 {
	switch wd {
	case time.Monday:
		return 1
	case time.Tuesday:
		return 2
	case time.Wednesday:
		return 3
	case time.Thursday:
		return 4
	case time.Friday:
		return 5
	case time.Saturday:
		return 6
	case time.Sunday:
		return 7
	default:
		return 0
	}
}

// normalizeDays собирает массив дней недели в map для быстрого lookup.
func normalizeDays(days []int16) map[int16]bool {
	m := make(map[int16]bool, len(days))
	for _, d := range days {
		m[d] = true
	}
	return m
}

// NormalizeAndValidateDays проверяет, что каждый элемент в диапазоне [1,7]
// (ISO-номера дней), убирает дубликаты и сортирует. Возвращает
// domain.ErrInvalidRequest при пустом результате или невалидных значениях.
func NormalizeAndValidateDays(days []int16) ([]int16, error) {
	set := map[int16]struct{}{}
	for _, d := range days {
		if d < 1 || d > 7 {
			return nil, domain.ErrInvalidRequest
		}
		set[d] = struct{}{}
	}

	result := make([]int16, 0, len(set))
	for d := range set {
		result = append(result, d)
	}

	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	if len(result) == 0 {
		return nil, domain.ErrInvalidRequest
	}

	return result, nil
}
