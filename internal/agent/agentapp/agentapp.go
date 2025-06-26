package agentapp

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/agent/sender"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
)

type App struct {
	monitoring *monitoring.MonitoringStats
	sender     *sender.Sender
	config     config.AgentConfig
}

func New() *App {
	cfg := config.LoadAgentCfg(os.Args[1:]...)
	monitor := monitoring.New(cfg.PollInterval)
	logger.Initialize(cfg.LogLvl, cfg.IsDev)
	cfgString := fmt.Sprintf("%+v\n", cfg)
	logger.Log.Info(cfgString)

	return &App{
		monitoring: monitor,
		sender:     sender.New(cfg, monitor),
		config:     cfg,
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
