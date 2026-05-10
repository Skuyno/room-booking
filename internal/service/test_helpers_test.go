package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

type userRepoStub struct {
	createFn     func(ctx context.Context, user domain.User) (domain.User, error)
	getByEmailFn func(ctx context.Context, email string) (domain.User, error)
}

func (s userRepoStub) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if s.createFn != nil {
		return s.createFn(ctx, user)
	}
	return user, nil
}

func (s userRepoStub) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if s.getByEmailFn != nil {
		return s.getByEmailFn(ctx, email)
	}
	return domain.User{}, nil
}

type roomRepoStub struct {
	createFn func(ctx context.Context, room domain.Room) error
	listFn   func(ctx context.Context) ([]domain.Room, error)
	existsFn func(ctx context.Context, roomID uuid.UUID) (bool, error)
}

func (s roomRepoStub) Create(ctx context.Context, room domain.Room) error {
	if s.createFn != nil {
		return s.createFn(ctx, room)
	}
	return nil
}

func (s roomRepoStub) List(ctx context.Context) ([]domain.Room, error) {
	if s.listFn != nil {
		return s.listFn(ctx)
	}
	return nil, nil
}

func (s roomRepoStub) Exists(ctx context.Context, roomID uuid.UUID) (bool, error) {
	if s.existsFn != nil {
		return s.existsFn(ctx, roomID)
	}
	return false, nil
}

type scheduleRepoStub struct {
	createFn      func(ctx context.Context, schedule domain.Schedule) error
	getByRoomIDFn func(ctx context.Context, roomID uuid.UUID) (domain.Schedule, error)
}

func (s scheduleRepoStub) Create(ctx context.Context, schedule domain.Schedule) error {
	if s.createFn != nil {
		return s.createFn(ctx, schedule)
	}
	return nil
}

func (s scheduleRepoStub) GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, error) {
	if s.getByRoomIDFn != nil {
		return s.getByRoomIDFn(ctx, roomID)
	}
	return domain.Schedule{}, domain.ErrScheduleNotFound
}

type bookingRepoStub struct {
	createFn                        func(ctx context.Context, booking domain.Booking) error
	getByIDFn                       func(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error)
	cancelFn                        func(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error)
	listByUserFutureFn              func(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error)
	listAllFn                       func(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
	listActiveStartsByRoomBetweenFn func(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]time.Time, error)
}

func (s bookingRepoStub) Create(ctx context.Context, booking domain.Booking) error {
	if s.createFn != nil {
		return s.createFn(ctx, booking)
	}
	return nil
}

func (s bookingRepoStub) GetByID(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, bookingID)
	}
	return domain.Booking{}, nil
}

func (s bookingRepoStub) Cancel(ctx context.Context, bookingID uuid.UUID) (domain.Booking, error) {
	if s.cancelFn != nil {
		return s.cancelFn(ctx, bookingID)
	}
	return domain.Booking{}, nil
}

func (s bookingRepoStub) ListByUserFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error) {
	if s.listByUserFutureFn != nil {
		return s.listByUserFutureFn(ctx, userID, now)
	}
	return nil, nil
}

func (s bookingRepoStub) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	if s.listAllFn != nil {
		return s.listAllFn(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (s bookingRepoStub) ListActiveStartsByRoomBetween(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]time.Time, error) {
	if s.listActiveStartsByRoomBetweenFn != nil {
		return s.listActiveStartsByRoomBetweenFn(ctx, roomID, from, to)
	}
	return nil, nil
}

func mustUUID(value string) uuid.UUID {
	return uuid.MustParse(value)
}
