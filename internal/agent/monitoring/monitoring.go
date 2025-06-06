// Package monitoring собирает метрики с помощью библиотек gopsutil, и runtime.MemStats.

package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type StatsMap map[string]float64

type MonitoringStats struct {
	PollSignal   chan any
	runtimeStats StatsMap
	gopsStats    StatsMap
	PollInterval int
	mu           sync.RWMutex
}

func New(interval int) *MonitoringStats {
	return &MonitoringStats{
		PollInterval: interval,
		PollSignal:   make(chan any),
		runtimeStats: make(StatsMap),
		gopsStats:    make(StatsMap),
	}
}

func (m *MonitoringStats) saveGopsStats() {
	cpuPercent, err := cpu.Percent(time.Duration(0), true)
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := make(StatsMap, len(cpuPercent)+2)
	defer func() {
		m.gopsStats = stats
	}()
	if err != nil {
		logger.Log.Error(err.Error())
	} else {
		cpuName := "CPUutilization"
		for i, val := range cpuPercent {
			key := fmt.Sprintf("%s%d", cpuName, i+1)
			stats[key] = val
		}
	}

	diskInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("saveGopsStats: can't get diskinfo. " + err.Error())
		return
	}

	stats["TotalMemory"] = float64(diskInfo.Total)
	stats["FreeMemory"] = float64(diskInfo.Free)
}

func (m *MonitoringStats) saveRuntimeStats() {
	var runtimeStats runtime.MemStats

	runtime.ReadMemStats(&runtimeStats)
	m.mu.Lock()
	m.runtimeStats = StatsMap{
		"Alloc":         float64(runtimeStats.Alloc),
		"BuckHashSys":   float64(runtimeStats.BuckHashSys),
		"Frees":         float64(runtimeStats.Frees),
		"GCCPUFraction": float64(runtimeStats.GCCPUFraction),
		"GCSys":         float64(runtimeStats.GCSys),
		"HeapAlloc":     float64(runtimeStats.HeapAlloc),
		"HeapIdle":      float64(runtimeStats.HeapIdle),
		"HeapInuse":     float64(runtimeStats.HeapInuse),
		"HeapObjects":   float64(runtimeStats.HeapObjects),
		"HeapReleased":  float64(runtimeStats.HeapReleased),
		"HeapSys":       float64(runtimeStats.HeapSys),
		"LastGC":        float64(runtimeStats.LastGC),
		"Lookups":       float64(runtimeStats.Lookups),
		"MCacheInuse":   float64(runtimeStats.MCacheInuse),
		"MCacheSys":     float64(runtimeStats.MCacheSys),
		"MSpanInuse":    float64(runtimeStats.MSpanInuse),
		"MSpanSys":      float64(runtimeStats.MSpanSys),
		"Mallocs":       float64(runtimeStats.Mallocs),
		"NextGC":        float64(runtimeStats.NextGC),
		"NumForcedGC":   float64(runtimeStats.NumForcedGC),
		"NumGC":         float64(runtimeStats.NumGC),
		"OtherSys":      float64(runtimeStats.OtherSys),
		"PauseTotalNs":  float64(runtimeStats.PauseTotalNs),
		"StackInuse":    float64(runtimeStats.StackInuse),
		"StackSys":      float64(runtimeStats.StackSys),
		"Sys":           float64(runtimeStats.Sys),
		"TotalAlloc":    float64(runtimeStats.TotalAlloc),
	}
	m.mu.Unlock()
	m.PollSignal <- struct{}{}
}

func (m *MonitoringStats) copyStats(stats StatsMap) StatsMap {
	copy := make(StatsMap, len(stats))
	for key, val := range stats {
		copy[key] = val
	}

	return copy
}

func (m *MonitoringStats) GetRuntimeStats() StatsMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.copyStats(m.runtimeStats)
}

func (m *MonitoringStats) GetGopsStats() StatsMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.copyStats(m.gopsStats)
}

func (m *MonitoringStats) CreateTicker(
	ctx context.Context, wg *sync.WaitGroup, tfunc func(),
) {
	ticker := time.NewTicker(time.Duration(m.PollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stop monitoring ticker")
			ticker.Stop()
			wg.Done()
			return
		case <-ticker.C:
			tfunc()
		}
	}
}

func (m *MonitoringStats) Start(
	ctx context.Context, wg *sync.WaitGroup,
) {
	logger.Log.Info("Start get metrics")
	m.saveGopsStats()
	m.saveRuntimeStats()

	wg.Add(2)
	go m.CreateTicker(ctx, wg, m.saveRuntimeStats)
	go m.CreateTicker(ctx, wg, m.saveGopsStats)
}
