package main

import (
	"sync"

	"github.com/LekcRg/metrics/internal/agent/metrics"
	"github.com/LekcRg/metrics/internal/agent/runtimeMonitoring"
)

var pollInterval = 2
var reportInterval = 10

func main() {
	var monitor map[string]float64
	var wg sync.WaitGroup
	wg.Add(2)
	go runtimeMonitoring.Start(&monitor, pollInterval)
	// wg.Done() // to stop
	go metrics.StartSending(&monitor, reportInterval)
	// wg.Done() // to stop
	wg.Wait()
}
