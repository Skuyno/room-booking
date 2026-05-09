package http

import (
	"context"

	api "github.com/Skuyno/room-booking/api/gen"
)

// PostBookingsCreate обрабатывает POST /bookings/create — создание брони.
// Авторизация: только роль user (admin не может бронировать). userId
// берётся из JWT, не из тела — клиент не может «забронировать за другого».
func (h *Handler) PostBookingsCreate(
	ctx context.Context,
	request api.PostBookingsCreateRequestObject,
) (api.PostBookingsCreateResponseObject, error) {
	if request.Body == nil {
		return api.PostBookingsCreate400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "invalid request"),
		), nil
	}

	actor, err := actorFromContext(ctx)
	if err != nil {
		return api.PostBookingsCreate401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	createConferenceLink := false
	if request.Body.CreateConferenceLink != nil {
		createConferenceLink = *request.Body.CreateConferenceLink
	}

	booking, err := h.bookingService.Create(
		ctx,
		actor.UserID,
		actor.Role,
		request.Body.RoomId,
		request.Body.StartAt,
		createConferenceLink,
	)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 400:
			return api.PostBookingsCreate400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 403:
			return api.PostBookingsCreate403JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 404:
			return api.PostBookingsCreate404JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 409:
			return api.PostBookingsCreate409JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.PostBookingsCreate500JSONResponse(
				newInternalErrorResponse("failed to create booking"),
			), nil
		}
	}

	respBooking := bookingToAPI(booking)
	return api.PostBookingsCreate201JSONResponse{
		Booking: &respBooking,
	}, nil
}

// GetBookingsMy обрабатывает GET /bookings/my — листинг будущих броней
// текущего пользователя. Авторизация: только роль user. Прошлые брони в
// выдачу не попадают.
func (h *Handler) GetBookingsMy(
	ctx context.Context,
	_ api.GetBookingsMyRequestObject,
) (api.GetBookingsMyResponseObject, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return api.GetBookingsMy401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	bookings, err := h.bookingService.ListMy(ctx, actor.UserID, actor.Role)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 403:
			return api.GetBookingsMy403JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.GetBookingsMy500JSONResponse(
				newInternalErrorResponse("failed to list user bookings"),
			), nil
		}
	}

	respBookings := bookingsToAPI(bookings)
	return api.GetBookingsMy200JSONResponse{
		Bookings: &respBookings,
	}, nil
}

// GetBookingsList обрабатывает GET /bookings/list — листинг всех броней
// с пагинацией. Авторизация: только роль admin. Дефолты page=1, pageSize=20.
func (h *Handler) GetBookingsList(
	ctx context.Context,
	request api.GetBookingsListRequestObject,
) (api.GetBookingsListResponseObject, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return api.GetBookingsList401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	page := 1
	if request.Params.Page != nil {
		page = *request.Params.Page
	}

	pageSize := 20
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	bookings, total, err := h.bookingService.ListAll(ctx, actor.Role, page, pageSize)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 400:
			return api.GetBookingsList400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 403:
			return api.GetBookingsList403JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.GetBookingsList500JSONResponse(
				newInternalErrorResponse("failed to list bookings"),
			), nil
		}
	}

	respBookings := bookingsToAPI(bookings)
	respPagination := api.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}

	return api.GetBookingsList200JSONResponse{
		Bookings:   &respBookings,
		Pagination: &respPagination,
	}, nil
}

// PostBookingsBookingIdCancel обрабатывает POST /bookings/{bookingId}/cancel —
// отмену брони. Авторизация: только роль user, и только своя бронь
// (проверка владельца — внутри сервиса). Идемпотентно.
func (h *Handler) PostBookingsBookingIdCancel(
	ctx context.Context,
	request api.PostBookingsBookingIdCancelRequestObject,
) (api.PostBookingsBookingIdCancelResponseObject, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return api.PostBookingsBookingIdCancel401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "unauthorized"),
		), nil
	}

	booking, err := h.bookingService.Cancel(ctx, actor.UserID, actor.Role, request.BookingId)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 403:
			return api.PostBookingsBookingIdCancel403JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 404:
			return api.PostBookingsBookingIdCancel404JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.PostBookingsBookingIdCancel500JSONResponse(
				newInternalErrorResponse("failed to cancel booking"),
			), nil
		}
	}

	respBooking := bookingToAPI(booking)
	return api.PostBookingsBookingIdCancel200JSONResponse{
		Booking: &respBooking,
	}, nil
}
