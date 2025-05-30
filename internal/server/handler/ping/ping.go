package ping

import (
	"context"
	"io"
	"net/http"
)

// PingService описывает интерфейс для проверки доступности БД.
type PingService interface {
	Ping(ctx context.Context) error
}

// Ping — хендлер, который проверяет доступность БД через PingService.
// Возвращает статусы 200 при успехе или 500 при ошибке.
func Ping(p PingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := p.Ping(r.Context())
		if err != nil {
			http.Error(w, "Internal error 500", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
		io.WriteString(w, "200 OK")
	}
}
