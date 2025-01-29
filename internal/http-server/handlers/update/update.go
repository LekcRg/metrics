package update

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/LekcRg/metrics/internal/storage"
)

type database interface {
	UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
	UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
	GetAllCounter() (storage.CounterCollection, error)
	GetAllGouge() (storage.GaugeCollection, error)
}

func sendErrorValue(w http.ResponseWriter, err error, errorTextType string) {
	textErr := fmt.Sprintf(
		"Bad request: incorrect value. The counter value must be %s\n\n%s",
		errorTextType, err.Error(),
	)
	http.Error(w, textErr, http.StatusBadRequest)
}

func New(db database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)

			return
		}

		contentType := r.Header.Get("Content-type")
		strippedPath := strings.TrimSuffix(r.URL.Path[1:], "/")
		parsedPath := strings.Split(strippedPath, "/")

		if contentType != "text/plain" {
			http.Error(w, "Incorrect content-type", http.StatusNotFound)
			return
		}

		if len(parsedPath) < 4 {
			if len(parsedPath) == 2 {
				http.Error(w, "Not found 404", http.StatusNotFound)
			} else {
				http.Error(w, "Bad request 400", http.StatusBadRequest)
			}
			return
		}

		reqType := parsedPath[1]
		reqName := parsedPath[2]
		reqValue := parsedPath[3]

		if reqType == "counter" {
			value, err := strconv.ParseInt(reqValue, 0, 64)
			if err != nil {
				sendErrorValue(w, err, "int64")
				return
			}
			db.UpdateCounter(reqName, storage.Counter(value))
		} else if reqType == "gauge" {
			value, err := strconv.ParseFloat(reqValue, 64)
			if err != nil {
				sendErrorValue(w, err, "float64")
				return
			}
			db.UpdateGauge(reqName, storage.Gauge(value))
		} else {
			http.Error(
				w,
				"Bad request: incorrect type. The type must be a counter or a gauge",
				http.StatusBadRequest,
			)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
}
