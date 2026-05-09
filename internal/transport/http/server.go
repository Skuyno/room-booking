// Package http собирает HTTP-сервер: маршрутизацию (chi), middleware,
// strict-server из oapi-codegen и адаптацию доменных ошибок к ответам API.
package http

import (
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	api "github.com/Skuyno/room-booking/api/gen"
	"github.com/Skuyno/room-booking/internal/auth"
	appmw "github.com/Skuyno/room-booking/internal/transport/http/middleware"
)

// NewServer собирает HTTP-обработчик с подключёнными middleware и
// сгенерированным strict-server. Порядок middleware имеет значение:
//
//  1. Recover — ловит панику и превращает в 500 (должен быть первым,
//     иначе паника может пройти мимо других middleware).
//  2. Logging — логирует каждый запрос с финальным статусом и временем.
//  3. Auth — валидирует JWT и кладёт claims в контекст (пропускает
//     публичные пути типа /_info и /login).
//
// strict-server получает обработчики ошибок: невалидный запрос →
// единый 400 INVALID_REQUEST, ошибка формирования ответа → 500.
func NewServer(handler *Handler, jwtManager *auth.JWTManager) stdhttp.Handler {
	r := chi.NewRouter()

	r.Use(appmw.Recover)
	r.Use(appmw.Logging)
	r.Use(appmw.Auth(jwtManager))

	strictHandler := api.NewStrictHandlerWithOptions(
		handler,
		nil,
		api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  handleStrictRequestError,
			ResponseErrorHandlerFunc: handleStrictResponseError,
		},
	)

	return api.HandlerFromMux(strictHandler, r)
}
