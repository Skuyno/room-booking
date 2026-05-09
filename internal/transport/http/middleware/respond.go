package middleware

import (
	"encoding/json"
	"net/http"
)

// writeJSONError отправляет стандартное JSON-тело ошибки. Используется
// middleware-ами, которые отвечают сами (Recover, Auth) — в обычных
// хендлерах ответы формируются через сгенерированные strict-server типы.
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
