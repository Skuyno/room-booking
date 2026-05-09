package http

import (
	"context"

	api "github.com/Skuyno/room-booking/api/gen"
)

// GetInfo обрабатывает GET /_info — health/info-эндпоинт без авторизации.
// Всегда возвращает {"status":"ok"}.
func (h *Handler) GetInfo(
	_ context.Context,
	_ api.GetInfoRequestObject,
) (api.GetInfoResponseObject, error) {
	return api.GetInfo200JSONResponse{
		Status: "ok",
	}, nil
}

// PostDummyLogin обрабатывает POST /dummyLogin — выдачу тестового JWT по
// роли. Этот эндпоинт обязателен по ТЗ, но в проде должен быть отключён.
func (h *Handler) PostDummyLogin(
	ctx context.Context,
	request api.PostDummyLoginRequestObject,
) (api.PostDummyLoginResponseObject, error) {
	if request.Body == nil {
		return api.PostDummyLogin400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "invalid request"),
		), nil
	}

	token, err := h.authService.DummyLogin(ctx, string(request.Body.Role))
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 400:
			return api.PostDummyLogin400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.PostDummyLogin500JSONResponse(
				newInternalErrorResponse("failed to generate token"),
			), nil
		}
	}

	return api.PostDummyLogin200JSONResponse{
		Token: token,
	}, nil
}

// PostRegister обрабатывает POST /register — регистрацию пользователя.
// Доступен без авторизации; роль передаётся в теле и определяет права
// будущих JWT-токенов.
func (h *Handler) PostRegister(
	ctx context.Context,
	request api.PostRegisterRequestObject,
) (api.PostRegisterResponseObject, error) {
	if request.Body == nil {
		return api.PostRegister400JSONResponse(
			newErrorResponse(api.INVALIDREQUEST, "invalid request"),
		), nil
	}

	user, err := h.authService.Register(
		ctx,
		string(request.Body.Email),
		request.Body.Password,
		string(request.Body.Role),
	)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 400:
			return api.PostRegister400JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		default:
			return api.PostRegister500JSONResponse(
				newInternalErrorResponse("failed to register user"),
			), nil
		}
	}

	respUser := userToAPI(user)
	return api.PostRegister201JSONResponse{
		User: &respUser,
	}, nil
}

// PostLogin обрабатывает POST /login — авторизацию по email/паролю.
// Возвращает 401 для любой неудачной попытки (нет юзера, не подходит
// пароль, формат email невалиден), чтобы не сигналить наружу о
// существовании email.
func (h *Handler) PostLogin(
	ctx context.Context,
	request api.PostLoginRequestObject,
) (api.PostLoginResponseObject, error) {
	if request.Body == nil {
		return api.PostLogin401JSONResponse(
			newErrorResponse(api.UNAUTHORIZED, "invalid credentials"),
		), nil
	}

	token, err := h.authService.Login(
		ctx,
		string(request.Body.Email),
		request.Body.Password,
	)
	if err != nil {
		apiErr := mapDomainError(err)

		switch apiErr.Status {
		case 401:
			return api.PostLogin401JSONResponse(
				newErrorResponse(apiErr.Code, apiErr.Message),
			), nil
		case 400:
			return api.PostLogin401JSONResponse(
				newErrorResponse(api.UNAUTHORIZED, "invalid credentials"),
			), nil
		default:
			return api.PostLogin500JSONResponse(
				newInternalErrorResponse("failed to login"),
			), nil
		}
	}

	return api.PostLogin200JSONResponse{
		Token: token,
	}, nil
}
