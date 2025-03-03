package value

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/LekcRg/metrics/internal/server/services/metric"
)

func Get(s metric.MetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqType := chi.URLParam(r, "type")
		reqName := chi.URLParam(r, "name")

		res, err := s.GetMetric(reqName, reqType)

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, res)
	}
}
