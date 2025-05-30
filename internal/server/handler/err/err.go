package err

import "net/http"

// ErrorBadRequest — хендлер, который возвращает 400-ую ошибку.
func ErrorBadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "400 Bad request", http.StatusBadRequest)
}
