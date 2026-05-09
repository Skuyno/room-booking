package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Skuyno/room-booking/internal/domain"
)

// RoomService отвечает за создание и листинг переговорок.
type RoomService struct {
	repo RoomRepository
}

// NewRoomService собирает сервис из репозитория комнат.
func NewRoomService(repo RoomRepository) *RoomService {
	return &RoomService{repo: repo}
}

// Create валидирует вход (непустое имя, положительная вместимость) и
// создаёт переговорку с новым UUID. Авторизация (только admin) делается
// слоем выше, в HTTP-хендлере.
func (s *RoomService) Create(
	ctx context.Context,
	name string,
	description *string,
	capacity *int,
) (domain.Room, error) {
	if name == "" {
		return domain.Room{}, domain.ErrInvalidRequest
	}
	if capacity != nil && *capacity <= 0 {
		return domain.Room{}, domain.ErrInvalidRequest
	}

	room := domain.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}

	if err := s.repo.Create(ctx, room); err != nil {
		return domain.Room{}, err
	}

	return room, nil
}

// List возвращает все переговорки в порядке возрастания created_at.
func (s *RoomService) List(ctx context.Context) ([]domain.Room, error) {
	return s.repo.List(ctx)
}
