package http

import (
	"context"

	api "github.com/Skuyno/room-booking/api/gen"
)

// GetRoomsRoomIdSlotsList обрабатывает GET /rooms/{roomId}/slots/list — выдачу
// свободных слотов комнаты на дату. Доступен любой авторизованной роли.
//
// Если у комнаты нет расписания, сервис возвращает пустой список — это не
// ошибка, а нормальное состояние.
func (h *Handler) GetRoomsRoomIdSlotsList(
	ctx context.Context,
	request api.GetRoomsRoomIdSlotsListRequestObject,
) (api.GetRoomsRoomIdSlotsListResponseObject, error) {
	_, err := actorFromContext(ctx)
	if err != nil {
		return api.GetRoomsRoomIdSlotsList401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	slots, err := h.slotService.ListAvailableByRoomAndDate(
		ctx,
		request.RoomId,
		request.Params.Date.Time,
	)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 400:
			return api.GetRoomsRoomIdSlotsList400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 404:
			return api.GetRoomsRoomIdSlotsList404JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.GetRoomsRoomIdSlotsList500JSONResponse(
				newInternalErrorResponse("failed to list slots"),
			), nil
		}
	}

	respSlots := slotsToAPI(slots)
	return api.GetRoomsRoomIdSlotsList200JSONResponse{
		Slots: &respSlots,
	}, nil
}
