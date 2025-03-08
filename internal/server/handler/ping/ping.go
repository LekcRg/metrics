package ping

import (
	"io"
	"net/http"

	"github.com/LekcRg/metrics/internal/server/services/dbping"
)

func Ping(p dbping.PingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := p.Ping()
		if err != nil {
			http.Error(w, "Internal error 500", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
		io.WriteString(w, "200 OK")
	}
}
