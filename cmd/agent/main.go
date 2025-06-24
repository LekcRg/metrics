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

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func PrintBuildInfo() {
	fmt.Println("Build version: " + buildVersion)
	fmt.Println("Build date: " + buildDate)
	fmt.Println("Build commit: " + buildCommit)
}

func exit(exited chan any,
	s *sender.Sender, m *monitoring.MonitoringStats) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigChan
	logger.Log.Info("Stopping, wait end of requests")
	m.Shutdown()
	s.Shutdown()

	exited <- true
}

func main() {
	PrintBuildInfo()
	config := config.LoadAgentCfg(os.Args[1:]...)
	logger.Initialize(config.LogLvl, config.IsDev)
	cfgString := fmt.Sprintf("%+v\n", config)
	logger.Log.Info(cfgString)

	var wg sync.WaitGroup
	ctx := context.Background()

	monitor := monitoring.New(config.PollInterval)
	send := sender.New(config, monitor)

	send.Start(ctx, &wg)
	monitor.Start(ctx, &wg)

	exited := make(chan any, 1)
	go exit(exited, send, monitor)
	wg.Wait()

	<-exited
	logger.Log.Info("Buy, 👋!")
}
