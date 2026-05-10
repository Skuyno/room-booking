package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

// BookingService отвечает за создание, отмену и листинг броней. Бизнес-правила:
//   - бронировать может только роль user (admin не может);
//   - отменять можно только свою бронь;
//   - бронь в прошлом запрещена;
//   - startAt должен попадать в активное расписание комнаты и лежать на
//     30-минутной сетке (валидируется через IsStartAtWithinSchedule);
//   - повторная отмена идемпотентна.
type BookingService struct {
	scheduleRepo ScheduleRepository
	bookingRepo  BookingRepository
	nowFn        func() time.Time
}

// NewBookingService собирает сервис из репозиториев расписаний и броней.
// nowFn по умолчанию возвращает текущее время в UTC; в тестах можно подменять.
func NewBookingService(
	scheduleRepo ScheduleRepository,
	bookingRepo BookingRepository,
) *BookingService {
	return &BookingService{
		scheduleRepo: scheduleRepo,
		bookingRepo:  bookingRepo,
		nowFn:        func() time.Time { return time.Now().UTC() },
	}
}

// Create создаёт активную бронь на 30-минутный слот в комнате roomID,
// начиная с startAt. Возможные ошибки:
//   - domain.ErrForbidden — роль не user;
//   - domain.ErrInvalidRequest — startAt в прошлом;
//   - domain.ErrSlotNotFound — у комнаты нет расписания, либо startAt вне
//     расписания (неподходящий день, время или невыровненная сетка);
//   - domain.ErrSlotAlreadyBooked — пара (roomID, startAt) уже занята.
//
// При createConferenceLink=true в бронь записывается мок-ссылка вида
// https://conference.mock/<uuid>. Реальной интеграции с конф-сервисом нет.
func (s *BookingService) Create(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	roomID uuid.UUID,
	startAt time.Time,
	createConferenceLink bool,
) (domain.Booking, error) {
	if role != domain.RoleUser {
		return domain.Booking{}, domain.ErrForbidden
	}

	startAt = startAt.UTC()

	if startAt.Before(s.nowFn()) {
		return domain.Booking{}, domain.ErrInvalidRequest
	}

	schedule, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			return domain.Booking{}, domain.ErrSlotNotFound
		}
		return domain.Booking{}, err
	}

	if !IsStartAtWithinSchedule(schedule, startAt) {
		return domain.Booking{}, domain.ErrSlotNotFound
	}

	endAt := startAt.Add(SlotDuration)

	var conferenceLink *string
	if createConferenceLink {
		link := "https://conference.mock/" + uuid.NewString()
		conferenceLink = &link
	}

	booking := domain.Booking{
		ID:             uuid.New(),
		RoomID:         roomID,
		UserID:         userID,
		StatusID:       domain.BookingStatusIDActive,
		Status:         domain.BookingStatusActive,
		StartAt:        startAt,
		EndAt:          endAt,
		ConferenceLink: conferenceLink,
	}

	if err = s.bookingRepo.Create(ctx, booking); err != nil {
		return domain.Booking{}, err
	}

	return booking, nil
}

// Cancel помечает бронь cancelled. Только владелец брони (по userID) может
// её отменить — иначе domain.ErrForbidden. Повторный вызов на уже отменённой
// брони тоже успешен и возвращает её актуальное состояние.
func (s *BookingService) Cancel(
	ctx context.Context,
	userID uuid.UUID,
	role string,
	bookingID uuid.UUID,
) (domain.Booking, error) {
	if role != domain.RoleUser {
		return domain.Booking{}, domain.ErrForbidden
	}

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return domain.Booking{}, err
	}

	if booking.UserID != userID {
		return domain.Booking{}, domain.ErrForbidden
	}

	return s.bookingRepo.Cancel(ctx, bookingID)
}

// ListMy возвращает будущие брони пользователя. Прошедшие брони и брони
// других пользователей в выдачу не попадают.
func (s *BookingService) ListMy(
	ctx context.Context,
	userID uuid.UUID,
	role string,
) ([]domain.Booking, error) {
	if role != domain.RoleUser {
		return nil, domain.ErrForbidden
	}

	return s.bookingRepo.ListByUserFuture(ctx, userID, s.nowFn())
}

// ListAll возвращает страницу всех броней (admin-only) с total для
// пагинации. Невалидные page/pageSize нормализуются (1/20 по умолчанию),
// pageSize > 100 → ErrInvalidRequest.
func (s *BookingService) ListAll(
	ctx context.Context,
	role string,
	page int,
	pageSize int,
) ([]domain.Booking, int, error) {
	if role != domain.RoleAdmin {
		return nil, 0, domain.ErrForbidden
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		return nil, 0, domain.ErrInvalidRequest
	}

	return s.bookingRepo.ListAll(ctx, page, pageSize)
}
