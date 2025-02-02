package main

import (
	"github.com/LekcRg/metrics/internal/server/storage/memStorage"
	"net/http"
	"os"

	"github.com/LekcRg/metrics/internal/server/router"
)

func main() {
	storage, err := memStorage.New()
	if err != nil {
		os.Exit(1)
	}
	router := router.NewRouter(storage)
	http.ListenAndServe(":8080", router)
}
