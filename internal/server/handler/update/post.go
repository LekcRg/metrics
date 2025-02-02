package update

import (
	"fmt"
	"github.com/LekcRg/metrics/internal/server/storage"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

//	type database interface {
//		UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
//		UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
//		GetGaugeByName(name string) (storage.Gauge, error)
//		GetCounterByName(name string) (storage.Counter, error)
//		// GetAllCounter() (storage.CounterCollection, error)
//		// GetAllGouge() (storage.GaugeCollection, error)
//	}
// type database interface {
// 	UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
// 	UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
// 	// GetGaugeByName(name string) (storage.Gauge, error)
// 	// GetCounterByName(name string) (storage.Counter, error)
// 	// GetAllCounter() (storage.CounterCollection, error)
// 	// GetAllGouge() (storage.GaugeCollection, error)
// }

func sendErrorValue(w http.ResponseWriter, err error, errorTextType string) {
	textErr := fmt.Sprintf(
		"Bad request: incorrect value. The counter value must be %s\n\n%s",
		errorTextType, err.Error(),
	)
	http.Error(w, textErr, http.StatusBadRequest)
}

func Post(db database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")

		if contentType != "text/plain" {
			http.Error(w, "Incorrect content-type", http.StatusBadRequest)
			return
		}

		reqType := chi.URLParam(r, "type")
		reqName := chi.URLParam(r, "name")
		reqValue := chi.URLParam(r, "value")

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
		io.WriteString(w, "success")
	}
}
