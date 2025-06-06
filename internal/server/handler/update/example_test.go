package update

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

func ExamplePost() {
	s := &MockMetricUpdater{}
	s.On("UpdateMetric", mock.Anything, "example", "gauge", "1.1").Return(nil)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", Post(s))
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/gauge/example/1.1", nil)
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
	// success
	// 200 OK
}

func ExamplePostJSON() {
	s := &MockMetricUpdater{}
	var gaugeVal storage.Gauge = 12.34
	metric := models.Metrics{
		ID:    "example",
		MType: "gauge",
		Value: &gaugeVal,
	}

	jsonSend, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("create jsonSend err")
	}

	s.On("UpdateMetricJSON", mock.Anything, metric).Return(metric, nil)

	router := chi.NewRouter()
	router.Post("/update", PostJSON(s))
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update", bytes.NewReader(jsonSend))
	if err != nil {
		panic("error create request\n" + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil || res == nil {
		panic("error send req\n" + err.Error())
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

func ExamplePostMany() {
	s := &MockMetricUpdater{}
	var gaugeVal storage.Gauge = 12.34
	var counterVal storage.Counter = 11
	metrics := []models.Metrics{
		{
			ID:    "example1",
			MType: "gauge",
			Value: &gaugeVal,
		},
		{
			ID:    "example2",
			MType: "counter",
			Delta: &counterVal,
		},
	}

	jsonSend, err := json.Marshal(metrics)
	if err != nil {
		fmt.Println("create jsonSend err")
	}

	s.On("UpdateMany", mock.Anything, metrics).Return(nil)

	router := chi.NewRouter()
	router.Post("/updates", PostMany(s, ""))
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/updates", bytes.NewReader(jsonSend))
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
	// Success
	// 200 OK
}
