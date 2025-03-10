package router

import (
	"github.com/LekcRg/metrics/internal/cgzip"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/handler/home"
	"github.com/LekcRg/metrics/internal/server/services"
	"github.com/go-chi/chi/v5"
)

func NewRouter(metricService services.MetricService) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)

	// or just use middleware by NYTimes
	r.Use(cgzip.GzipHandle)
	r.Use(cgzip.GzipBody)

	r.Get("/", home.Get(metricService))
	UpdateRoutes(r, metricService)
	ValueRoutes(r, metricService)

	return r
}
