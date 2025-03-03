package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	server := &http.Server{
		Addr:    addrFlag,
		Handler: router,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		store.Save(updateService, fileStoragePath)
		if err := server.Close(); err != nil {
			logger.Log.Fatal("HTTP close error")
		}
	}()

	err = server.ListenAndServe()
	logger.Log.Fatal(err.Error())
}
