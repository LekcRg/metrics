package main

import (
	"net/http"
	"os"

	"github.com/LekcRg/metrics/internal/http-server/handlers/update"
	"github.com/LekcRg/metrics/internal/storage/memStorage"
)

func main() {
	storage, err := memStorage.New()
	if err != nil {
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", update.New(storage))

	http.ListenAndServe(":8080", mux)
}
