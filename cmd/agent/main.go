package main

import (
	"sync"

	"github.com/LekcRg/metrics/internal/agent/metrics"
	"github.com/LekcRg/metrics/internal/agent/monitoring"
)

func main() {
	parseFlags()
	var monitor map[string]float64
	var wg sync.WaitGroup
	wg.Add(2)

	readySignal := make(chan bool)
	go monitoring.Start(&monitor, pollInterval, readySignal)
	// wg.Done() // to stop
	// wait get monitoring
	<-readySignal
	go metrics.StartSending(&monitor, reportInterval, addrFlag, https)
	// wg.Done() // to stop
	wg.Wait()
}
