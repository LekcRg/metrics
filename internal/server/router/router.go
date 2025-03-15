package router

import (
	"github.com/LekcRg/metrics/internal/cgzip"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/handler/home"
	"github.com/LekcRg/metrics/internal/server/handler/ping"
	"github.com/LekcRg/metrics/internal/server/services/dbping"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/go-chi/chi/v5"
)

func NewRouter(metricService metric.MetricService, pingService dbping.PingService) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)

	// or just use middleware by NYTimes
	r.Use(cgzip.GzipHandle)
	r.Use(cgzip.GzipBody)

	r.Get("/", home.Get(metricService))
	r.Get("/ping", ping.Ping(pingService))
	UpdateRoutes(r, metricService)
	ValueRoutes(r, metricService)

	return r
}
