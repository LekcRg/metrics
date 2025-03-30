package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/cgzip"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/retry"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type Sender struct {
	url       string
	config    config.AgentConfig
	monitor   *monitoring.MonitoringStats
	countSent int
	jobs      chan []byte
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
	}
}

func (s *Sender) postRequest(ctx context.Context, body []byte) error {
	req, err := cgzip.GetGzippedReq(ctx, s.url, body)
	if err != nil {
		logger.Log.Error("Error while getting gzipped request")
		return err
	}

	if s.config.Key != "" {
		sha := crypto.GenerateSHA256(body, s.config.Key)
		req.Header.Set("HashSHA256", sha)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	var resp *http.Response

	err = retry.Retry(ctx, func() error {
		resp, err = client.Do(req)
		if err != nil {
			return err
		}

		defer func() {
			if resp != nil {
				resp.Body.Close()
			}
		}()
		return nil
	})

	if err != nil {
		logger.Log.Error("Error making http request")
		return err
	}

	if resp.StatusCode > 299 {
		logger.Log.Warn("Server answered with status code: " +
			strconv.Itoa(resp.StatusCode))
		return fmt.Errorf("invalid status code")
	}

	s.countSent++
	logger.Log.Info("Request sent. Total: " + strconv.Itoa(s.countSent) + " requests")

	return nil
}

func (s *Sender) postRequestWorker(ctx context.Context) {
	for data := range s.jobs {
		s.postRequest(ctx, data)
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
			case <-ctx.Done():
				logger.Log.Info("Stop sending metrics")
				close(s.jobs)
				ticker.Stop()
				wg.Done()
				return
			case <-s.monitor.PollSignal:
				s.sendPollCount(ctx)
			case <-ticker.C:
				s.sendAllMetrics(ctx)
			}
		}
	}()
}
