package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/LekcRg/metrics/internal/cgzip"
	"github.com/LekcRg/metrics/internal/common"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func postRequest(ctx context.Context, url string, body []byte, key string) error {
	req, err := cgzip.GetGzippedReq(url, body)
	if err != nil {
		logger.Log.Error("Error while getting gzipped request")
		return err
	}

	if key != "" {
		sha := common.GenerateSHA256(body, key)
		req.Header.Set("HashSHA256", sha)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	var resp *http.Response

	err = common.Retry(ctx, func() error {
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
		logger.Log.Warn("Server answered with status code: " + strconv.Itoa(resp.StatusCode))
		return fmt.Errorf("invalid status code")
	}

	return nil
}

func sendMetrics(ctx context.Context, list []models.Metrics, baseURL string, key string) error {
	jsonBody, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return postRequest(ctx, baseURL, jsonBody, key)
}

func getRandomValue() storage.Gauge {
	randomValueLeft := storage.Gauge(rand.IntN(99999))
	randomValueRight := storage.Gauge(float64(rand.IntN(999)) / 1000)

	return randomValueLeft + randomValueRight
}

func generateJSON(
	mType string, name string, value *storage.Gauge, delta *storage.Counter,
) models.Metrics {
	return models.Metrics{
		MType: mType,
		ID:    name,
		Value: value,
		Delta: delta,
	}
}

func sendAllMetrics(ctx context.Context, monitor *map[string]float64, baseURL string, countSent *int, key string) {
	*countSent++
	countRequests := 1
	list := []models.Metrics{}
	pollCountVal := storage.Counter(0)
	for key, value := range *monitor {
		countRequests++
		sendVal := storage.Gauge(value)
		list = append(list, generateJSON("gauge", key, &sendVal, nil))
		pollCountVal++
	}

	randomVal := getRandomValue()
	list = append(list, generateJSON("gauge", "RandomValue", &randomVal, nil))
	list = append(list, generateJSON("counter", "PollCount", nil, &pollCountVal))

	err := sendMetrics(ctx, list, baseURL, key)

	if err != nil {
		*countSent--
		return
	}
	logger.Log.Info(strconv.Itoa(*countSent) + " time sent. Now was " + strconv.Itoa(countRequests) + " requests")
}

func StartSending(ctx context.Context, wg *sync.WaitGroup, monitor *map[string]float64, config config.AgentConfig) {
	defer wg.Done()
	baseURL := config.Addr + "/updates/"
	if config.IsHTTPS {
		baseURL = "https://" + baseURL
	} else {
		baseURL = "http://" + baseURL
	}

	countSent := 0
	sendAllMetrics(ctx, monitor, baseURL, &countSent, config.Key)

	ticker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stop send metrics")
			return
		case <-ticker.C:
			sendAllMetrics(ctx, monitor, baseURL, &countSent, config.Key)
		}
	}
}
