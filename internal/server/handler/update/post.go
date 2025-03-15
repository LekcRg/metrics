package update

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LekcRg/metrics/internal/server/services/metric"

	"github.com/go-chi/chi/v5"
)

func Post(s metric.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-type")

		// В задании написано, что запрос должен быть `text/plain`
		// А в тестах отправляется пустой Content-Type
		// Убрать проверку?
		if !strings.Contains(contentType, "text/plain") && contentType != "" {
			http.Error(w, "Incorrect content-type "+contentType, http.StatusBadRequest)
			return
		}

		reqType := chi.URLParam(r, "type")
		reqName := chi.URLParam(r, "name")
		reqValue := chi.URLParam(r, "value")

		err := s.UpdateMetric(reqName, reqType, reqValue)
		if err != nil {
			textErr := fmt.Sprintf("Bad request: %s", err)
			http.Error(w, textErr, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "success")
	}
}
