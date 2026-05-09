package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/domain"
)

// APIError — промежуточная структура для маппинга domain-ошибки в
// HTTP-ответ. Хендлеры на основе Status выбирают конкретный
// сгенерированный response-тип (api.PostXxx400JSONResponse и т.п.).
type APIError struct {
	Status  int
	Code    api.ErrorResponseErrorCode
	Message string
}

// newErrorResponse собирает ErrorResponse — стандартное тело ошибок API.
func newErrorResponse(code api.ErrorResponseErrorCode, message string) api.ErrorResponse {
	var resp api.ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	return resp
}

// newInternalErrorResponse собирает InternalErrorResponse для 500-ответов.
// Он отличается от ErrorResponse только тем, что код всегда INTERNAL_ERROR.
func newInternalErrorResponse(message string) api.InternalErrorResponse {
	var resp api.InternalErrorResponse
	resp.Error.Code = string(api.INTERNALERROR)
	resp.Error.Message = message
	return resp
}

// mapDomainError переводит domain-ошибку в HTTP-статус, код и сообщение.
// Это единственное место, где доменный мир встречается с HTTP — все
// хендлеры пользуются этим маппером, а не дублируют switch у себя.
//
// Ошибки сравниваются через errors.Is, поэтому wrap через fmt.Errorf("...: %w", err)
// в нижних слоях не ломает классификацию.
func mapDomainError(err error) APIError {
	switch {
	case errors.Is(err, domain.ErrInvalidRequest):
		return APIError{
			Status:  stdhttp.StatusBadRequest,
			Code:    api.INVALIDREQUEST,
			Message: "invalid request",
		}
	case errors.Is(err, domain.ErrEmailTaken):
		return APIError{
			Status:  stdhttp.StatusBadRequest,
			Code:    api.INVALIDREQUEST,
			Message: "email is already taken",
		}
	case errors.Is(err, domain.ErrInvalidCredentials):
		return APIError{
			Status:  stdhttp.StatusUnauthorized,
			Code:    api.UNAUTHORIZED,
			Message: "invalid credentials",
		}
	case errors.Is(err, domain.ErrForbidden):
		return APIError{
			Status:  stdhttp.StatusForbidden,
			Code:    api.FORBIDDEN,
			Message: "forbidden",
		}
	case errors.Is(err, domain.ErrRoomNotFound):
		return APIError{
			Status:  stdhttp.StatusNotFound,
			Code:    api.ROOMNOTFOUND,
			Message: "room not found",
		}
	case errors.Is(err, domain.ErrSlotNotFound):
		return APIError{
			Status:  stdhttp.StatusNotFound,
			Code:    api.SLOTNOTFOUND,
			Message: "slot not found",
		}
	case errors.Is(err, domain.ErrBookingNotFound):
		return APIError{
			Status:  stdhttp.StatusNotFound,
			Code:    api.BOOKINGNOTFOUND,
			Message: "booking not found",
		}
	case errors.Is(err, domain.ErrSlotAlreadyBooked):
		return APIError{
			Status:  stdhttp.StatusConflict,
			Code:    api.SLOTALREADYBOOKED,
			Message: "slot is already booked",
		}
	case errors.Is(err, domain.ErrScheduleExists):
		return APIError{
			Status:  stdhttp.StatusConflict,
			Code:    api.SCHEDULEEXISTS,
			Message: "schedule for this room already exists and cannot be changed",
		}
	default:
		return APIError{
			Status:  stdhttp.StatusInternalServerError,
			Code:    api.INTERNALERROR,
			Message: "internal server error",
		}
	}
}

// handleStrictRequestError вызывается oapi-codegen, когда не удалось
// распарсить или провалидировать входящий запрос. Возвращает единый
// 400 INVALID_REQUEST, чтобы детали парсера не утекали наружу.
func handleStrictRequestError(w stdhttp.ResponseWriter, _ *stdhttp.Request, _ error) {
	resp := newErrorResponse(api.INVALIDREQUEST, "invalid request")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(resp)
}

// handleStrictResponseError вызывается oapi-codegen, когда не удалось
// сериализовать ответ хендлера. Это путь «никогда не должно случиться»,
// поэтому отвечаем 500 без подробностей.
func handleStrictResponseError(w stdhttp.ResponseWriter, _ *stdhttp.Request, _ error) {
	resp := newInternalErrorResponse("internal server error")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stdhttp.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(resp)
}
