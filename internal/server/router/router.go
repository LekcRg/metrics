package router

import (
	"github.com/LekcRg/metrics/internal/server/handler/home"
	"github.com/LekcRg/metrics/internal/server/storage/memStorage"
	"github.com/go-chi/chi/v5"
)

func NewRouter(storage *memStorage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Get("/", home.Get(storage))
	UpdateRoutes(r, storage)
	ValueRotes(r, storage)

	return r
}
