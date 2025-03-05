package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LekcRg/metrics/internal/agent/metrics"
	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
)

func exit(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()
}

func main() {
	config := config.LoadAgentCfg()
	logger.Initialize(config.LogLvl, config.IsDev)
	var monitor map[string]float64
	var wg sync.WaitGroup
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())

	readySignal := make(chan bool)
	logger.Log.Info("Start get metrics")
	go monitoring.Start(ctx, &wg, &monitor, config.PollInterval, readySignal)

	// wait get monitoring
	<-readySignal

	logger.Log.Info("Start sending metrics")
	go metrics.StartSending(ctx, &wg, &monitor, config.ReportInterval, config.Addr, config.IsHTTPS)
	go exit(cancel)
	wg.Wait()
	logger.Log.Info("Buy, 👋!")
}
