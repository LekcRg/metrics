package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LekcRg/metrics/internal/common"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/services/metric"
)

func validateSHA256(w http.ResponseWriter, r *http.Request,
	body []byte, config config.ServerConfig) error {
	if config.Key != "" {
		headerSHA := r.Header.Get("HashSHA256")
		if headerSHA == "" {
			http.Error(w, "Bad request: empty HashSHA256", http.StatusBadRequest)
			return fmt.Errorf("empty hash")
		}

		sha := common.GenerateSHA256(body, config.Key)
		if sha != headerSHA {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return fmt.Errorf("hash is not correct")
		}
	}

	return nil
}

func validateAndGetBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	contentType := r.Header.Get("Content-type")

	if !strings.Contains(contentType, "application/json") && contentType != "" {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil, fmt.Errorf("incorrect content-type")
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func PostJSON(s metric.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := validateAndGetBody(w, r)
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

		newMetric, err := s.UpdateMetricJSON(r.Context(), parsedBody)
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

func PostMany(s metric.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := validateAndGetBody(w, r)
		if err != nil {
			logger.Log.Error("/update: body reading error " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = validateSHA256(w, r, body, s.Config)
		if err != nil {
			logger.Log.Error("Error while validating SHA256 hash " + err.Error())
			return
		}

		var parsedBody []models.Metrics
		err = json.Unmarshal(body, &parsedBody)
		if err != nil {
			logger.Log.Error("error while unmarshal json")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = s.UpdateMany(r.Context(), parsedBody)
		if err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Success"))
	}
}
