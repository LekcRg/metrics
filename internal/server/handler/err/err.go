package err

import "net/http"

func ErrorBadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "400 Bad request", http.StatusBadRequest)
}
