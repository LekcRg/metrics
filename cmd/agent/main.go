package main

import (
	"sync"

	"github.com/LekcRg/metrics/internal/agent/metrics"
	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/logger"
)

func main() {
	parseFlags()
	logger.Initialize(logLvl, isDev)
	var monitor map[string]float64
	var wg sync.WaitGroup
	wg.Add(2)

	readySignal := make(chan bool)
	logger.Log.Info("Start get metrics")
	go monitoring.Start(&monitor, pollInterval, readySignal)

	// wait get monitoring
	<-readySignal

	logger.Log.Info("Start sending metrics")
	go metrics.StartSending(&monitor, reportInterval, addrFlag, https)
	wg.Wait()
}
