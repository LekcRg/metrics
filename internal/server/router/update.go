package router

import (
	"net/http"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/handler/err"
	"github.com/LekcRg/metrics/internal/server/handler/update"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/go-chi/chi/v5"
)

func UpdateRoutes(
	r chi.Router, metricService metric.MetricService, cfg config.ServerConfig,
) {
	r.Route("/update", func(r chi.Router) {
		r.Post("/", update.PostJSON(&metricService))
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", http.NotFound)
			r.Post("/{name}", err.ErrorBadRequest)
			r.Post("/{name}/{value}", update.Post(&metricService))
		})
	})
	r.Post("/updates/", update.PostMany(&metricService, cfg.Key))
}
