package router

import (
	"github.com/LekcRg/metrics/internal/server/handler"
	"github.com/LekcRg/metrics/internal/server/storage/memStorage"
	"github.com/go-chi/chi/v5"
)

func NewRouter(storage *memStorage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Get("/", handler.HomeGet(storage))
	UpdateRoutes(r, storage)

	return r
}
