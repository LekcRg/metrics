package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/agent/sender"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
)

//go:generate go run ../prebuild/prebuild.go -version 0.20

func exit(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()
}

func main() {
	config := config.LoadAgentCfg()
	logger.Initialize(config.LogLvl, config.IsDev)
	cfgString := fmt.Sprintf("%+v\n", config)
	logger.Log.Info(cfgString)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	monitor := monitoring.New(config.PollInterval)
	send := sender.New(config, monitor)

	send.Start(ctx, &wg)
	monitor.Start(ctx, &wg)

	go exit(cancel)
	wg.Wait()
	logger.Log.Info("Buy, 👋!")
}
