package serverapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/router"
	"github.com/LekcRg/metrics/internal/server/services/dbping"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
	"github.com/LekcRg/metrics/internal/server/storage/postgres"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type App struct {
	server *http.Server
	store  *store.Store
	db     storage.Storage
	config config.ServerConfig
}

func initDB(ctx context.Context, cfg config.ServerConfig) (storage.Storage, error) {
	var db storage.Storage
	var err error

	if cfg.DatabaseDSN != "" {
		logger.Log.Info("create pg storage")
		db, err = postgres.NewPostgres(ctx, cfg)
	} else {
		logger.Log.Info("Create memstorage")
		db, err = memstorage.New()
	}

	return db, err
}

func initRouter(cfg config.ServerConfig, db storage.Storage, store metric.Store) chi.Router {
	logger.Log.Info("Create dbping service")
	ping := dbping.NewPing(db, cfg)

	logger.Log.Info("Create metric service")
	metricService := metric.NewMetricsService(db, cfg, store)

	logger.Log.Info("Create router")
	return router.NewRouter(router.NewRouterArgs{
		MetricService: *metricService,
		PingService:   *ping,
		Cfg:           cfg,
	})
}

func New(ctx context.Context, wg *sync.WaitGroup) (*App, error) {
	config := config.LoadServerCfg(os.Args[1:]...)
	logger.Initialize(config.LogLvl, config.IsDev)
	cfgString := fmt.Sprintf("%+v\n", config)
	logger.Log.Info(cfgString)

	db, err := initDB(ctx, config)
	if err != nil {
		return nil, err
	}

	logger.Log.Info("Create store service")
	store := store.NewStore(db, config)

	router := initRouter(config, db, store)

	if config.Restore {
		store.Restore(ctx)
	}

	if !config.SyncSave && config.StoreInterval > 0 {
		wg.Add(1)
		logger.Log.Info("Start saving store")
		go store.StartSaving(ctx, wg)
	}

	server := &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	return &App{
		config: config,
		server: server,
		store:  store,
		db:     db,
	}, nil
}

func (app *App) Start() {
	err := app.server.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Log.Error(err.Error())
	}
}

func (app *App) Stop(ctx context.Context) {
	if app.server != nil {
		if err := app.server.Shutdown(ctx); err != nil {
			switch err {
			case http.ErrServerClosed:
				logger.Log.Info("Server was already closed")
			case context.DeadlineExceeded:
				logger.Log.Error("Shutdown timeout exceeded, forcing close", zap.Error(err))
				app.server.Close()
			default:
				logger.Log.Error("Server shutdown error", zap.Error(err))
			}
		}
	}
	if app.store != nil {
		err := app.store.Save(context.Background())
		if err != nil {
			logger.Log.Error("Error while saving store")
		}
	}
	app.db.Close()
}
