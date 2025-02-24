package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func postRequest(url string, body io.Reader) {
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		logger.Log.Error("Error making http request")
		return
	}
	resp.Body.Close()
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

	postRequest(baseURL, bytes.NewBuffer(jsonBody))
}

func getRandomValue() storage.Gauge {
	randomValueLeft := storage.Gauge(rand.IntN(99999))
	randomValueRight := storage.Gauge(float64(rand.IntN(999)) / 1000)

	return randomValueLeft + randomValueRight
}

func StartSending(monitor *map[string]float64, interval int, addr string, https bool) {
	baseURL := addr + "/update"
	if https {
		baseURL = "https://" + baseURL
	} else {
		baseURL = "http://" + baseURL
	}

	countSent := 0
	for {
		countSent++
		for key, value := range *monitor {
			sendVal := storage.Gauge(value)
			sendMetric("gauge", key, &sendVal, nil, baseURL)
			pollCountVal := storage.Counter(1)
			sendMetric("counter", "PollCount", nil, &pollCountVal, baseURL)
		}

		randomVal := getRandomValue()
		sendMetric("gauge", "RandomValue", &randomVal, nil, baseURL)
		fmt.Printf("%d time sent\n", countSent)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
