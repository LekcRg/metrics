package router

import (
	"github.com/LekcRg/metrics/internal/cgzip"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/handler/home"
	"github.com/LekcRg/metrics/internal/server/handler/ping"
	"github.com/LekcRg/metrics/internal/server/services/dbping"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/go-chi/chi/v5"
)

type NewRouterArgs struct {
	MetricService metric.MetricService
	PingService   dbping.PingService
	Cfg           config.ServerConfig
}

func NewRouter(args NewRouterArgs) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)

	// or just use middleware by NYTimes
	r.Use(cgzip.GzipHandle)
	r.Use(cgzip.GzipBody)

	r.Get("/", home.Get(&args.MetricService))
	r.Get("/ping", ping.Ping(args.PingService))
	UpdateRoutes(r, args.MetricService, args.Cfg)
	ValueRoutes(r, args.MetricService)

	return r
}
