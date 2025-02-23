package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/models"
	"github.com/LekcRg/metrics/internal/server/services"
)

func PostJSON(s services.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("adslkjfalskdjf")
		contentType := r.Header.Get("Content-type")

		if !strings.Contains(contentType, "application/json") && contentType != "" {
			http.Error(w, "Incorrect content-type "+contentType, http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("/update: body reading error")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var parsedBody models.Metrics
		err = json.Unmarshal(body, &parsedBody)
		if err != nil {
			logger.Log.Error("/update: json parsing error\n\n" + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Printf("%+v\n", parsedBody)

		newMetric, err := s.UpdateMetricJSON(parsedBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := json.Marshal(newMetric)
		if err != nil {
			logger.Log.Error("/update: Error while generate json response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}
