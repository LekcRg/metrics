package update

import (
	"github.com/LekcRg/metrics/internal/server/storage"
)

type database interface {
	UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
	UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
	GetGaugeByName(name string) (storage.Gauge, error)
	GetCounterByName(name string) (storage.Counter, error)
}

// func sendErrorValue(w http.ResponseWriter, err error, errorTextType string) {
// 	textErr := fmt.Sprintf(
// 		"Bad request: incorrect value. The counter value must be %s\n\n%s",
// 		errorTextType, err.Error(),
// 	)
// 	http.Error(w, textErr, http.StatusBadRequest)
// }

// func NewPostHandler(db database) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		contentType := r.Header.Get("Content-type")

// 		if contentType != "text/plain" {
// 			http.Error(w, "Incorrect content-type", http.StatusNotFound)
// 			return
// 		}

// 		reqType := chi.URLParam(r, "type")
// 		reqName := chi.URLParam(r, "name")
// 		reqValue := chi.URLParam(r, "value")

// 		if reqType == "counter" {
// 			value, err := strconv.ParseInt(reqValue, 0, 64)
// 			if err != nil {
// 				sendErrorValue(w, err, "int64")
// 				return
// 			}
// 			db.UpdateCounter(reqName, storage.Counter(value))
// 		} else if reqType == "gauge" {
// 			value, err := strconv.ParseFloat(reqValue, 64)
// 			if err != nil {
// 				sendErrorValue(w, err, "float64")
// 				return
// 			}
// 			db.UpdateGauge(reqName, storage.Gauge(value))
// 		}

// 		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 		w.WriteHeader(http.StatusOK)
// 		io.WriteString(w, "success")
// 	}
// }

// func NewGetHandler(db database) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		reqType := chi.URLParam(r, "type")
// 		reqName := chi.URLParam(r, "name")

// 		var (
// 			resVal string
// 			err    error
// 		)

// 		if reqType == "counter" {
// 			var val storage.Counter
// 			val, err = db.GetCounterByName(reqName)
// 			resVal = fmt.Sprintf("%d", val)
// 		} else if reqType == "gauge" {
// 			var val storage.Gauge
// 			val, err = db.GetGaugeByName(reqName)
// 			resVal = strconv.FormatFloat(float64(val), 'f', -1, 64)
// 		}

// 		if err != nil {
// 			http.Error(w, "Not found", http.StatusNotFound)
// 			return
// 		}

// 		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 		w.WriteHeader(http.StatusOK)
// 		io.WriteString(w, resVal)
// 	}
// }
