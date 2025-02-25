package value

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/LekcRg/metrics/internal/server/storage"

	"github.com/go-chi/chi/v5"
)

type database interface {
	GetGaugeByName(name string) (storage.Gauge, error)
	GetCounterByName(name string) (storage.Counter, error)
}

func Get(db database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqType := chi.URLParam(r, "type")
		reqName := chi.URLParam(r, "name")

		var (
			resVal string
			err    error
		)

		if reqType == "counter" {
			var val storage.Counter
			val, err = db.GetCounterByName(reqName)
			resVal = fmt.Sprintf("%d", val)
		} else if reqType == "gauge" {
			var val storage.Gauge
			val, err = db.GetGaugeByName(reqName)
			resVal = strconv.FormatFloat(float64(val), 'f', -1, 64)
		}

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, resVal)
	}
}
