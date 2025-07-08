package serverapp

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/grpcapi"
	"github.com/LekcRg/metrics/internal/server/router"
	"github.com/LekcRg/metrics/internal/server/services/dbping"
	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
	"github.com/LekcRg/metrics/internal/server/storage/postgres"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	server     *http.Server
	grpcServer *grpc.Server
	store      *store.Store
	db         storage.Storage
	config     config.ServerConfig
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

	logger.Log.Info("Create dbping service")
	ping := dbping.NewPing(db, config)

	logger.Log.Info("Create metric service")
	metricService := metric.NewMetricsService(db, config, store)

	logger.Log.Info("Create router")
	router := router.NewRouter(router.NewRouterArgs{
		MetricService: *metricService,
		PingService:   *ping,
		Cfg:           config,
	})

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

	grpcServer := grpcapi.NewServer(metricService, config)

	return &App{
		config:     config,
		server:     server,
		grpcServer: grpcServer,
		store:      store,
		db:         db,
	}, nil
}

func (app *App) Start(wg *sync.WaitGroup) {
	wg.Add(2)
	go func() {
		logger.Log.Info("Starting HTTP server")
		if err := app.server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Error(err.Error())
		}

		logger.Log.Info("HTTP server goroutine exited")
		wg.Done()
	}()
	go func() {
		listen, err := net.Listen("tcp", app.config.GRPCAddr)
		if err != nil {
			log.Fatal(err)
		}

		logger.Log.Info("Started GRPC server")
		if err := app.grpcServer.Serve(listen); err != nil {
			logger.Log.Info("Server GPRC server error", zap.Error(err))
		}

		logger.Log.Info("GRPC server goroutine exited")
		wg.Done()
	}()
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
	if app.grpcServer != nil {
		app.grpcServer.GracefulStop()
	}
	if app.store != nil {
		err := app.store.Save(context.Background())
		if err != nil {
			logger.Log.Error("Error while saving store")
		}
	}
	app.db.Close()
}
