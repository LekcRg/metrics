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

func sendMetrics(list []models.Metrics, baseURL string) {
	jsonBody, err := json.Marshal(list)
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

func sendAllMetrics(monitor *map[string]float64, baseURL string, countSent *int) {
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

	sendMetrics(list, baseURL)

	logger.Log.Info(strconv.Itoa(*countSent) + " time sent. Now was " + strconv.Itoa(countRequests) + " requests")
}

func StartSending(ctx context.Context, wg *sync.WaitGroup, monitor *map[string]float64, interval int, addr string, https bool) {
	baseURL := addr + "/updates/"
	if https {
		baseURL = "https://" + baseURL
	} else {
		baseURL = "http://" + baseURL
	}

	countSent := 0
	sendAllMetrics(monitor, baseURL, &countSent)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stop send metrics")
			wg.Done()
			return
		case <-ticker.C:
			sendAllMetrics(monitor, baseURL, &countSent)
		}
	}
}
