package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/router"
	"github.com/LekcRg/metrics/internal/server/services/dbping"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
	"github.com/LekcRg/metrics/internal/server/storage/postgres"
)

func exit(cancel context.CancelFunc, server *http.Server, store *store.Store, db storage.Storage) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()

	err := store.Save()
	if err != nil {
		logger.Log.Error("Error while saving store")
	}
	db.Close()
	if err := server.Close(); err != nil {
		logger.Log.Error("HTTP close error")
	}
}

func main() {
	config := config.LoadServerCfg()
	logger.Initialize(config.LogLvl, config.IsDev)
	cfgString := fmt.Sprintf("%+v\n", config)
	logger.Log.Info(cfgString)

	var db storage.Storage
	var err error

	if config.DatabaseDSN != "" {
		logger.Log.Info("create pg storage")
		db, err = postgres.NewPostgres(config)
	} else {
		logger.Log.Info("Create memstorage")
		db, err = memstorage.New()
	}
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	logger.Log.Info("Create dbping service")
	ping := dbping.NewPing(db, config)

	logger.Log.Info("Create store service")
	store := store.NewStore(db, config)

	logger.Log.Info("Create metric service")
	metricService := metric.NewMetricsService(db, config, store)

	logger.Log.Info("Create router")
	router := router.NewRouter(*metricService, *ping)

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

	go exit(cancel, server, store, db)

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Log.Error(err.Error())
	}

	wg.Wait()
	logger.Log.Info("Buy, 👋!")
}
