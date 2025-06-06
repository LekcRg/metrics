package value

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
)

func ExampleGet() {
	s := &MockMetricGetter{}
	s.On("GetMetric", mock.Anything, "example", "gauge").Return("12.34", nil)

	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", Get(s))
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/value/gauge/example", nil)
	if err != nil {
		panic("error create request")
	}
	res, err := ts.Client().Do(req)
	if err != nil || res == nil {
		panic("error send req")
	}

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return
	}

	fmt.Println(string(resBody))
	fmt.Println(res.Status)

	// Output:
	// 12.34
	// 200 OK
}

func ExamplePost() {
	s := &MockMetricGetter{}
	var gaugeVal storage.Gauge = 12.34
	metricSend := models.Metrics{
		ID:    "example",
		MType: "gauge",
	}
	metricWant := models.Metrics{
		ID:    "example",
		MType: "gauge",
		Value: &gaugeVal,
	}

	jsonSend, err := json.Marshal(metricSend)
	if err != nil {
		fmt.Println("create jsonSend err")
	}

	s.On("GetMetricJSON", mock.Anything, metricSend).Return(metricWant, nil)

	router := chi.NewRouter()
	router.Post("/value", Post(s))
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/value", bytes.NewReader(jsonSend))
	if err != nil {
		panic("error create request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil || res == nil {
		panic("error send req")
	}

	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return
	}

	fmt.Println(string(resBody))
	fmt.Println(res.Status)

	// Output:
	// {"value":12.34,"id":"example","type":"gauge"}
	// 200 OK
}
