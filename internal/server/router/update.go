package router

import (
	"github.com/LekcRg/metrics/internal/server/storage/memStorage"
	"net/http"

	"github.com/LekcRg/metrics/internal/server/handler/err"
	"github.com/LekcRg/metrics/internal/server/handler/update"
	"github.com/go-chi/chi/v5"
)

func UpdateRoutes(r chi.Router, storage *memStorage.MemStorage) chi.Router {
	r.Route("/update", func(r chi.Router) {
		r.Post("/", err.ErrorBadRequest)
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", http.NotFound)
			r.Post("/{name}", err.ErrorBadRequest)
			r.Get("/{name}", update.Get(storage))
			r.Post("/{name}/{value}", update.Post(storage))
		})
	})

	return r
}
