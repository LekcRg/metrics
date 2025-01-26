package update

import (
	"encoding/json"
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
		// remove / from start and end
		strippedPath := strings.TrimSuffix(r.URL.Path[1:], "/")
		parsedPath := strings.Split(strippedPath, "/")

		if contentType != "text/plain" {
			http.Error(w, "Incorrect content-type", http.StatusNotFound)
			return
		}

		if len(parsedPath) < 4 {
			http.Error(w, "Incorrect url", http.StatusNotFound)
			return
		}

		reqType := parsedPath[1]
		reqName := parsedPath[2]
		reqValue := parsedPath[3]

		if reqType != "counter" && reqType != "gauge" {
			http.Error(
				w,
				"Bad request: incorrect type. The type must be a counter or a gauge",
				http.StatusBadRequest,
			)
			return
		}

		var (
			jsonRes []byte
			jsonErr error
		)

		if reqType == "counter" {
			value, err := strconv.ParseInt(reqValue, 0, 64)
			if err != nil {
				sendErrorValue(w, err, "int64")
				return
			}
			// storage.CounterCollection[reqName] += counter(value)
			res, err := db.UpdateCounter(reqName, storage.Counter(value))
			if err != nil {
				http.Error(
					w,
					"Internal error",
					http.StatusInternalServerError,
				)
			}
			jsonRes, jsonErr = json.Marshal(res)
		} else if reqType == "gauge" {
			value, err := strconv.ParseFloat(reqValue, 64)
			if err != nil {
				sendErrorValue(w, err, "float64")
				return
			}
			// storage.GaugeCollection[reqName] = gauge(value)
			res, err := db.UpdateGauge(reqName, storage.Gauge(value))
			if err != nil {
				http.Error(
					w,
					"Internal error",
					http.StatusInternalServerError,
				)
			}
			jsonRes, jsonErr = json.Marshal(res)
		}

		if jsonErr != nil {
			http.Error(
				w,
				"can't provide a json. internal error",
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	}
}
