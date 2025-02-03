package main

import (
	"net/http"
	"os"

	"github.com/LekcRg/metrics/internal/server/storage/memStorage"

	"github.com/LekcRg/metrics/internal/server/router"
)

func main() {
	parseFlags()
	storage, err := memStorage.New()
	if err != nil {
		os.Exit(1)
	}
	router := router.NewRouter(storage)
	http.ListenAndServe(addrFlag, router)
}
