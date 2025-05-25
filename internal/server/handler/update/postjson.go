package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
)

func validateSHA256(w http.ResponseWriter, r *http.Request,
	body []byte, key string) error {
	if key != "" {
		headerSHA := r.Header.Get("HashSHA256")
		if headerSHA == "" {
			http.Error(w, "Bad request: empty HashSHA256", http.StatusBadRequest)
			return fmt.Errorf("empty hash")
		}

		sha := crypto.GenerateSHA256(body, key)
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
		http.Error(w, "Bad request: incorrect Content-Type", http.StatusBadRequest)
		return nil, fmt.Errorf("incorrect content-type")
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func PostJSON(s MetricUpdater) http.HandlerFunc {
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

func PostMany(s MetricUpdater, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := validateAndGetBody(w, r)
		if err != nil {
			logger.Log.Error("/update: body reading error " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = validateSHA256(w, r, body, key)
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
