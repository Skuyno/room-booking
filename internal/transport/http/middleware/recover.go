package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recover ловит panic в downstream-хендлерах и превращает в 500-ответ.
// Должен быть зарегистрирован первым в цепочке middleware, иначе паника
// может пройти мимо и обвалить весь процесс.
//
// Стектрейс пишется в лог через debug.Stack(); наружу ничего не утекает.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
				writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
