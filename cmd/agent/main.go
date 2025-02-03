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
	go monitoring.Start(&monitor, pollInterval)
	// wg.Done() // to stop
	go metrics.StartSending(&monitor, reportInterval, addrFlag, https)
	// wg.Done() // to stop
	wg.Wait()
}
