package ping

import (
	"context"
	"io"
	"net/http"
)

type PingService interface {
	Ping(ctx context.Context) error
}

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
