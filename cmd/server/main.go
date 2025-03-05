package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/router"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
)

func exit(cancel context.CancelFunc, server *http.Server, store *store.Store) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()

	err := store.Save()
	if err != nil {
		logger.Log.Error("Error while saving store")
	}
	if err := server.Close(); err != nil {
		logger.Log.Error("HTTP close error")
	}
}

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

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	if !config.SyncSave && config.StoreInterval > 0 {
		wg.Add(1)
		logger.Log.Info("Start saving store")
		go store.StartSaving(ctx, &wg)
	}

	server := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	go exit(cancel, server, store)

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Log.Error(err.Error())
	}

	wg.Wait()
	logger.Log.Info("Buy, 👋!")
}
