package metrics

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

var countSent = 0

func postRequest(url string) {
	_, err := http.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Printf("Url: %s\nError making http request: %s\n\n", url, err)
	}
}

func sendMetric(mType string, name string, value string, baseUrl string) {
	countSent++
	url := fmt.Sprintf(`%s/%s/%s/%s`, baseUrl, mType, name, value)
	postRequest(url)
}

func getRandomUrl() string {
	randomValueLeft := rand.IntN(99999)
	randomValueRight := rand.IntN(999)
	return fmt.Sprintf("%d.%d", randomValueLeft, randomValueRight)
}

func StartSending(monitor *map[string]float64, interval int, addr string, https bool) {
	baseUrl := addr + "/update"
	if https {
		baseUrl = "https://" + baseUrl
	} else {
		baseUrl = "http://" + baseUrl
	}

	pollCountUrl := baseUrl + "/counter/PollCount/1"
	for {
		for key, value := range *monitor {
			url := fmt.Sprintf(`%s/gauge/%s/%f`, baseUrl, key, value)
			postRequest(url)
			postRequest(pollCountUrl)
		}

		sendMetric("gauge", "RandomValue", getRandomUrl(), baseUrl)
		fmt.Printf("%d time sent\n", countSent)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
