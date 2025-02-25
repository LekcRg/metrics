package main

import (
	"net/http"
	"sync"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/services"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
	"github.com/LekcRg/metrics/internal/server/store"

	"github.com/LekcRg/metrics/internal/server/router"
)

func main() {
	parseFlags()

	logger.Initialize(logLvl, isDev)

	logger.Log.Info("Create storage")
	storage, err := memstorage.New()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Create metric service")
	updateService := services.NewMetricsService(storage)

	logger.Log.Info("Create router")
	router := router.NewRouter(updateService)

	logger.Log.Info("Start saving store")

	if restore {
		store.Restore(updateService, fileStoragePath)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go store.StartSaving(updateService, storeInterval, fileStoragePath)

	err = http.ListenAndServe(addrFlag, router)
	logger.Log.Fatal(err.Error())
}
