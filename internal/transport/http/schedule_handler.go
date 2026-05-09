package http

import (
	"context"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/domain"
)

// PostRoomsRoomIdScheduleCreate обрабатывает POST /rooms/{roomId}/schedule/create —
// создание расписания. Авторизация: только admin. roomId в path и body
// должен совпадать (защита от опечатки клиента).
func (h *Handler) PostRoomsRoomIdScheduleCreate(
	ctx context.Context,
	request api.PostRoomsRoomIdScheduleCreateRequestObject,
) (api.PostRoomsRoomIdScheduleCreateResponseObject, error) {
	if request.Body == nil {
		return api.PostRoomsRoomIdScheduleCreate400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "invalid request"),
		), nil
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return api.PostRoomsRoomIdScheduleCreate401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	if actor.Role != domain.RoleAdmin {
		return api.PostRoomsRoomIdScheduleCreate403JSONResponse(
			newErrorResponse(api.FORBIDDEN, "forbidden"),
		), nil
	}

	if request.Body.RoomId != request.RoomId {
		return api.PostRoomsRoomIdScheduleCreate400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "roomId in path and body must match"),
		), nil
	}

	days, ok := scheduleDaysToDomain(request.Body.DaysOfWeek)
	if !ok {
		return api.PostRoomsRoomIdScheduleCreate400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "invalid request"),
		), nil
	}

	schedule, err := h.scheduleService.Create(
		ctx,
		request.RoomId,
		days,
		request.Body.StartTime,
		request.Body.EndTime,
	)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 400:
			return api.PostRoomsRoomIdScheduleCreate400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 403:
			return api.PostRoomsRoomIdScheduleCreate403JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 404:
			return api.PostRoomsRoomIdScheduleCreate404JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 409:
			return api.PostRoomsRoomIdScheduleCreate409JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.PostRoomsRoomIdScheduleCreate500JSONResponse(
				newInternalErrorResponse("failed to create schedule"),
			), nil
		}
	}

	respSchedule := scheduleToAPI(schedule)
	return api.PostRoomsRoomIdScheduleCreate201JSONResponse{
		Schedule: &respSchedule,
	}, nil
}

// scheduleDaysToDomain переводит []int (как oapi-codegen генерирует
// числовые массивы) в []int16 (доменный формат) и валидирует диапазон [1,7].
// Возвращает (nil, false) при первом невалидном значении.
func scheduleDaysToDomain(daysOfWeek []int) ([]int16, bool) {
	days := make([]int16, 0, len(daysOfWeek))
	for _, day := range daysOfWeek {
		if day < 1 || day > 7 {
			return nil, false
		}
		days = append(days, int16(day))
	}

	return days, true
}
