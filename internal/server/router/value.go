package router

import (
	"github.com/LekcRg/metrics/internal/server/handler/value"
	"github.com/LekcRg/metrics/internal/server/storage/memStorage"
	"github.com/go-chi/chi/v5"
)

func ValueRotes(r chi.Router, storage *memStorage.MemStorage) {
	r.Route("/value", func(r chi.Router) {
		r.Route("/{type}", func(r chi.Router) {
			r.Get("/{name}", value.Get(storage))
		})
	})
}
