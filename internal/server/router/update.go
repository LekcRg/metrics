package router

import (
	"net/http"

	"github.com/LekcRg/metrics/internal/server/storage/memstorage"

	"github.com/LekcRg/metrics/internal/server/handler/err"
	"github.com/LekcRg/metrics/internal/server/handler/update"
	"github.com/go-chi/chi/v5"
)

func UpdateRoutes(r chi.Router, storage *memstorage.MemStorage) {
	r.Route("/update", func(r chi.Router) {
		r.Post("/", err.ErrorBadRequest)
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", http.NotFound)
			r.Post("/{name}", err.ErrorBadRequest)
			r.Post("/{name}/{value}", update.Post(storage))
		})
	})
}
