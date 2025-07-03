// Package sender получает метрики из monitoring.MonitoringStats и отправляет на сервер.
package sender

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/agent/req"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	"go.uber.org/zap"
)

type Sender struct {
	monitor   *monitoring.MonitoringStats
	jobs      chan []byte
	shutdown  chan bool
	url       string
	config    config.AgentConfig
	wg        sync.WaitGroup
	countSent int
}

func New(
	config config.AgentConfig, monitor *monitoring.MonitoringStats,
) *Sender {
	baseURL := config.Addr + "/updates/"
	if config.IsHTTPS {
		baseURL = "https://" + baseURL
	} else {
		baseURL = "http://" + baseURL
	}

	return &Sender{
		url:       baseURL,
		config:    config,
		monitor:   monitor,
		countSent: 0,
		jobs:      make(chan []byte),
		shutdown:  make(chan bool, 1),
	}
}

func (s *Sender) postRequestWorker(ctx context.Context) {
	for data := range s.jobs {
		s.wg.Add(1)
		func() {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
			defer func() {
				s.wg.Done()
				cancel()
			}()
			err := req.PostRequest(req.PostRequestArgs{
				Body:   data,
				Ctx:    ctx,
				Config: s.config,
				URL:    s.url,
			})

			if err != nil {
				logger.Log.Error("Error by PostRequest", zap.Error(err))
				return
			}

			s.countSent++
			logger.Log.Info("Request sent. Total: " + strconv.Itoa(s.countSent) + " requests")
		}()
	}
}

func (s *Sender) SendMetrics(ctx context.Context, list []models.Metrics) {
	jsonBody, err := json.Marshal(list)
	if err != nil {
		logger.Log.Error("Error while generate json")
		return
	}

	s.jobs <- jsonBody
}

func (s *Sender) getRandomValue() storage.Gauge {
	const intMax = 99999
	const fractMax = 999
	const fractDivisor = 1000

	intPart := storage.Gauge(rand.IntN(intMax))
	fractPart := storage.Gauge(float64(rand.IntN(999)) / fractDivisor)

	return intPart + fractPart
}

func (s *Sender) genMetricStruct(
	mType string, name string, value *storage.Gauge, delta *storage.Counter,
) models.Metrics {
	return models.Metrics{
		MType: mType,
		ID:    name,
		Value: value,
		Delta: delta,
	}
}

func (s *Sender) sendGaugeMetrics(
	ctx context.Context, stats monitoring.StatsMap,
) {
	list := []models.Metrics{}

	for key, value := range stats {
		sendVal := storage.Gauge(value)
		list = append(list, s.genMetricStruct("gauge", key, &sendVal, nil))
	}

	s.SendMetrics(ctx, list)
}

func (s *Sender) sendPollCount(ctx context.Context) {
	pollCountVal := storage.Counter(1)
	pollCountStruct := s.genMetricStruct("counter", "PollCount", nil, &pollCountVal)
	data := append([]models.Metrics{}, pollCountStruct)

	s.SendMetrics(ctx, data)
}

func (s *Sender) sendRandom(ctx context.Context) {
	randomVal := s.getRandomValue()
	randomStruct := s.genMetricStruct("gauge", "RandomValue", &randomVal, nil)
	data := append([]models.Metrics{}, randomStruct)

	s.SendMetrics(ctx, data)
}

func (s *Sender) sendAllMetrics(ctx context.Context) {
	// стоит объеденить в один запрос?
	s.sendGaugeMetrics(ctx, s.monitor.GetRuntimeStats())
	s.sendGaugeMetrics(ctx, s.monitor.GetGopsStats())
	s.sendRandom(ctx)
}

func (s *Sender) startWorkerPool(ctx context.Context) {
	for range s.config.RateLimit {
		go s.postRequestWorker(ctx)
	}
}

func (s *Sender) shutdownSender(wg *sync.WaitGroup) {
	logger.Log.Info("waiting...")
	s.wg.Wait()
	close(s.jobs)
	logger.Log.Info("Stop sending metrics")
	wg.Done()
}

func (s *Sender) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		<-s.monitor.PollSignal
		logger.Log.Info("Start sending metrics")

		s.startWorkerPool(ctx)
		ticker := time.NewTicker(time.Duration(s.config.ReportInterval) * time.Second)
		s.sendAllMetrics(ctx)
		s.sendPollCount(ctx)

		for {
			select {
			case <-s.shutdown:
				go s.shutdownSender(wg)
				ticker.Stop()
				return
			case <-s.monitor.PollSignal:
				s.sendPollCount(ctx)
			case <-ticker.C:
				s.sendAllMetrics(ctx)
			}
		}
	}()
}

func (s *Sender) Shutdown() {
	close(s.shutdown)
}
