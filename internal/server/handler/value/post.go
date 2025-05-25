package value

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/LekcRg/metrics/internal/models"

	"github.com/LekcRg/metrics/internal/logger"
)

func Post(s MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("/value: error while reading body")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		var parsedBody models.Metrics
		err = json.Unmarshal(body, &parsedBody)
		if err != nil {
			logger.Log.Error("/value: error while unmarshal json")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		res, err := s.GetMetricJSON(r.Context(), parsedBody)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		json, err := json.Marshal(res)
		if err != nil {
			logger.Log.Error("/value: error while marshal json")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}
