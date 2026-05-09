package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

// ScheduleService отвечает за создание расписания комнаты. После создания
// расписание неизменяемо — ни одного метода для редактирования нет.
// Слоты по расписанию вычисляются на лету в SlotService и не хранятся.
type ScheduleService struct {
	roomRepo     RoomRepository
	scheduleRepo ScheduleRepository
}

// NewScheduleService собирает сервис из репозиториев комнат и расписаний.
func NewScheduleService(
	roomRepo RoomRepository,
	scheduleRepo ScheduleRepository,
) *ScheduleService {
	return &ScheduleService{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
	}
}

// Create валидирует и сохраняет расписание для комнаты.
//
// Гарантии после успешного вызова:
//   - daysOfWeek нормализован (без дубликатов, отсортирован) и непуст;
//   - startTime/endTime — в формате HH:MM, минуты кратны 30, end > start;
//   - комната с roomID существует;
//   - запись в БД создана с уникальностью по room_id (повторная попытка
//     для той же комнаты приведёт к domain.ErrScheduleExists).
func (s *ScheduleService) Create(
	ctx context.Context,
	roomID uuid.UUID,
	daysOfWeek []int16,
	startTime string,
	endTime string,
) (domain.Schedule, error) {
	normalizedDays, err := NormalizeAndValidateDays(daysOfWeek)
	if err != nil {
		return domain.Schedule{}, err
	}
	if err := validateScheduleWindow(startTime, endTime); err != nil {
		return domain.Schedule{}, err
	}

	exists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return domain.Schedule{}, err
	}
	if !exists {
		return domain.Schedule{}, domain.ErrRoomNotFound
	}

	schedule := domain.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: normalizedDays,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return domain.Schedule{}, err
	}

	return schedule, nil
}

// validateScheduleWindow проверяет формат и согласованность временного окна:
// HH:MM, минуты кратны 30, end строго больше start. Возвращает
// domain.ErrInvalidRequest при любом нарушении.
func validateScheduleWindow(startTime, endTime string) error {
	startHour, startMinute, err := parseHHMM(startTime)
	if err != nil {
		return domain.ErrInvalidRequest
	}

	endHour, endMinute, err := parseHHMM(endTime)
	if err != nil {
		return domain.ErrInvalidRequest
	}

	if startMinute%30 != 0 || endMinute%30 != 0 {
		return domain.ErrInvalidRequest
	}

	startTotalMinutes := startHour*60 + startMinute
	endTotalMinutes := endHour*60 + endMinute
	if endTotalMinutes <= startTotalMinutes {
		return domain.ErrInvalidRequest
	}

	return nil
}
