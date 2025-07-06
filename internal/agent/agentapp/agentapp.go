package agentapp

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/agent/req"
	"github.com/LekcRg/metrics/internal/agent/sender"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
)

type App struct {
	monitoring *monitoring.MonitoringStats
	sender     *sender.Sender
	grpc       *req.GRPCClient
	config     config.AgentConfig
}

func New() *App {
	cfg := config.LoadAgentCfg(os.Args[1:]...)
	monitor := monitoring.New(cfg.PollInterval)
	logger.Initialize(cfg.LogLvl, cfg.IsDev)
	cfgString := fmt.Sprintf("%+v\n", cfg)
	logger.Log.Info(cfgString)

	var grpcCl *req.GRPCClient
	if cfg.IsGRPC {
		grpcCl = req.NewGRPCClient(cfg)
	}

	return &App{
		monitoring: monitor,
		sender:     sender.New(cfg, monitor, grpcCl),
		config:     cfg,
		grpc:       grpcCl,
	}
}

func (app *App) Start(ctx context.Context, wg *sync.WaitGroup) {
	app.sender.Start(ctx, wg)
	app.monitoring.Start(ctx, wg)
}

func (app *App) Stop() {
	app.sender.Shutdown()
	app.monitoring.Shutdown()
}
