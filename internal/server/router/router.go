package router

import (
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/handler/home"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
	"github.com/go-chi/chi/v5"
)

func NewRouter(storage *memstorage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Get("/", home.Get(storage))
	UpdateRoutes(r, storage)
	ValueRotes(r, storage)

	return r
}
