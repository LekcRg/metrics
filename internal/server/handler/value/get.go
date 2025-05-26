package value

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Get — хендлер для получения метрики по типу и имени из URL.
func Get(s MetricGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqType := chi.URLParam(r, "type")
		reqName := chi.URLParam(r, "name")

		res, err := s.GetMetric(r.Context(), reqName, reqType)

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, res)
	}
}
