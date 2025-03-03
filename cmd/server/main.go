package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"

	"github.com/LekcRg/metrics/internal/server/router"
)

func main() {
	config := config.LoadServerCfg()
	logger.Initialize(config.LogLvl, config.IsDev)

	logger.Log.Info("Create storage")
	storage, err := memstorage.New()
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Create store service")
	store := store.NewStore(storage, config)

	logger.Log.Info("Create metric service")
	updateService := metric.NewMetricsService(storage, config, store)

	logger.Log.Info("Create router")
	router := router.NewRouter(*updateService)

	if config.Restore {
		store.Restore()
	}

	if !config.SyncSave && config.StoreInterval > 0 {
		logger.Log.Info("Start saving store")
		var wg sync.WaitGroup
		wg.Add(1)
		go store.StartSaving()
	}

	server := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		err := store.Save()
		if err != nil {
			logger.Log.Error("Error while saving store")
		}
		if err := server.Close(); err != nil {
			logger.Log.Fatal("HTTP close error")
		}
	}()

	err = server.ListenAndServe()
	logger.Log.Fatal(err.Error())
}
