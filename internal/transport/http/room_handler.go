package http

import (
	"context"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/domain"
)

// GetRoomsList обрабатывает GET /rooms/list — листинг переговорок.
// Доступен любой авторизованной роли.
func (h *Handler) GetRoomsList(
	ctx context.Context,
	_ api.GetRoomsListRequestObject,
) (api.GetRoomsListResponseObject, error) {
	rooms, err := h.roomService.List(ctx)
	if err != nil {
		return api.GetRoomsList500JSONResponse(
			newInternalErrorResponse("failed to list rooms"),
		), nil
	}

	respRooms := roomsToAPI(rooms)
	return api.GetRoomsList200JSONResponse{
		Rooms: &respRooms,
	}, nil
}

// PostRoomsCreate обрабатывает POST /rooms/create — создание переговорки.
// Авторизация: только роль admin. Имя комнаты обрезается от пробелов
// перед валидацией и сохранением.
func (h *Handler) PostRoomsCreate(
	ctx context.Context,
	request api.PostRoomsCreateRequestObject,
) (api.PostRoomsCreateResponseObject, error) {
	if request.Body == nil || normalizedName(request.Body.Name) == "" {
		return api.PostRoomsCreate400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "invalid request"),
		), nil
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return api.PostRoomsCreate401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	if actor.Role != domain.RoleAdmin {
		return api.PostRoomsCreate403JSONResponse(
			newErrorResponse(api.FORBIDDEN, "forbidden"),
		), nil
	}

	room, err := h.roomService.Create(
		ctx,
		normalizedName(request.Body.Name),
		request.Body.Description,
		request.Body.Capacity,
	)
	if err != nil {
		apiErr := mapDomainError(err)
		if apiErr.Status == 400 {
			return api.PostRoomsCreate400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		}
		if apiErr.Status == 403 {
			return api.PostRoomsCreate403JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		}
		return api.PostRoomsCreate500JSONResponse(
			newInternalErrorResponse("failed to create room"),
		), nil
	}

	respRoom := roomToAPI(room)
	return api.PostRoomsCreate201JSONResponse{
		Room: &respRoom,
	}, nil
}
