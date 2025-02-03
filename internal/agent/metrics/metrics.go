package metrics

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

var countSent = 0

func postRequest(url string) {
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Printf("Url: %s\nError making http request: %s\n\n", url, err)
		return
	}
	resp.Body.Close()
}

func sendMetric(mType string, name string, value string, baseURL string) {
	countSent++
	url := fmt.Sprintf(`%s/%s/%s/%s`, baseURL, mType, name, value)
	postRequest(url)
}

func getRandomURL() string {
	randomValueLeft := rand.IntN(99999)
	randomValueRight := rand.IntN(999)
	return fmt.Sprintf("%d.%d", randomValueLeft, randomValueRight)
}

func StartSending(monitor *map[string]float64, interval int, addr string, https bool) {
	baseURL := addr + "/update"
	if https {
		baseURL = "https://" + baseURL
	} else {
		baseURL = "http://" + baseURL
	}

	pollCountURL := baseURL + "/counter/PollCount/1"
	for {
		for key, value := range *monitor {
			url := fmt.Sprintf(`%s/gauge/%s/%f`, baseURL, key, value)
			postRequest(url)
			postRequest(pollCountURL)
		}

		sendMetric("gauge", "RandomValue", getRandomURL(), baseURL)
		fmt.Printf("%d time sent\n", countSent)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
