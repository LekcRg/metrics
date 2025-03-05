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
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func postRequest(url string, body []byte) {
	req, err := cgzip.GetGzippedReq(url, body)
	if err != nil {
		logger.Log.Error("Error while geting gzipped request")
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error("Error making http request")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		logger.Log.Warn("Server answered with status code: " + strconv.Itoa(resp.StatusCode))
	}
}

func sendMetric(
	mType string, name string,
	value *storage.Gauge, delta *storage.Counter,
	baseURL string,
) {
	body := models.Metrics{
		ID:    name,
		MType: mType,
		Value: value,
		Delta: delta,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		fmt.Println()
	}

	postRequest(baseURL, jsonBody)
}

func getRandomValue() storage.Gauge {
	randomValueLeft := storage.Gauge(rand.IntN(99999))
	randomValueRight := storage.Gauge(float64(rand.IntN(999)) / 1000)

	return randomValueLeft + randomValueRight
}

func StartSending(ctx context.Context, wg *sync.WaitGroup, monitor *map[string]float64, interval int, addr string, https bool) {
	baseURL := addr + "/update"
	if https {
		baseURL = "https://" + baseURL
	} else {
		baseURL = "http://" + baseURL
	}

	countSent := 1
	randomVal := getRandomValue()
	sendMetric("gauge", "RandomValue", &randomVal, nil, baseURL)
	logger.Log.Info(strconv.Itoa(countSent) + " time sent")

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stop send metrics")
			wg.Done()
			return
		case <-ticker.C:
			countSent++
			for key, value := range *monitor {
				sendVal := storage.Gauge(value)
				sendMetric("gauge", key, &sendVal, nil, baseURL)
				pollCountVal := storage.Counter(1)
				sendMetric("counter", "PollCount", nil, &pollCountVal, baseURL)
			}

			randomVal := getRandomValue()
			sendMetric("gauge", "RandomValue", &randomVal, nil, baseURL)
			logger.Log.Info(strconv.Itoa(countSent) + " time sent")
		}
	}
}
