package router

import (
	"github.com/LekcRg/metrics/internal/server/handler/err"
	"github.com/LekcRg/metrics/internal/server/handler/value"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/go-chi/chi/v5"
)

func ValueRoutes(r chi.Router, metricService metric.MetricService) {
	r.Route("/value", func(r chi.Router) {
		r.Post("/", value.Post(metricService))
		r.Route("/{type:counter|gauge}", func(r chi.Router) {
			r.Get("/{name}", value.Get(metricService))
		})
		r.Get("/{type}/{name}", err.ErrorBadRequest)
	})
}
