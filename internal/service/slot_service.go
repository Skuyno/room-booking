package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

// SlotService отвечает за выдачу свободных слотов комнаты на конкретную
// дату. Слоты не материализуются в БД — они вычисляются из расписания
// функцией GenerateSlotsForDate, а занятые исключаются по результатам
// запроса к BookingRepository.
type SlotService struct {
	roomRepo     RoomRepository
	scheduleRepo ScheduleRepository
	bookingRepo  BookingRepository
}

// NewSlotService собирает сервис из репозиториев комнат, расписаний и броней.
func NewSlotService(
	roomRepo RoomRepository,
	scheduleRepo ScheduleRepository,
	bookingRepo BookingRepository,
) *SlotService {
	return &SlotService{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
		bookingRepo:  bookingRepo,
	}
}

// ListAvailableByRoomAndDate возвращает свободные 30-минутные слоты комнаты
// на указанную дату.
//
// Шаги работы:
//  1. Проверка существования комнаты → domain.ErrRoomNotFound, если её нет.
//  2. Чтение расписания комнаты. Если расписания нет — возвращается
//     пустой список (это нормальное состояние, не ошибка).
//  3. Генерация всех возможных слотов на дату через GenerateSlotsForDate.
//  4. Чтение start_at активных броней комнаты в [dayStart, dayEnd).
//  5. Вычитание занятых из сгенерированных.
//
// Дата интерпретируется в UTC: всё, что лежит вне UTC-суток, отбрасывается.
func (s *SlotService) ListAvailableByRoomAndDate(
	ctx context.Context,
	roomID uuid.UUID,
	date time.Time,
) ([]domain.Slot, error) {
	exists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrRoomNotFound
	}

	schedule, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			return []domain.Slot{}, nil
		}
		return nil, err
	}

	slots, err := GenerateSlotsForDate(schedule, date)
	if err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		return []domain.Slot{}, nil
	}

	dayStart := truncateToDayUTC(date)
	dayEnd := dayStart.Add(24 * time.Hour)

	bookedStarts, err := s.bookingRepo.ListActiveStartsByRoomBetween(ctx, roomID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	booked := make(map[time.Time]struct{}, len(bookedStarts))
	for _, t := range bookedStarts {
		booked[t.UTC()] = struct{}{}
	}

	result := make([]domain.Slot, 0, len(slots))
	for _, slot := range slots {
		if _, isBooked := booked[slot.StartAt]; isBooked {
			continue
		}
		result = append(result, slot)
	}

	return result, nil
}
