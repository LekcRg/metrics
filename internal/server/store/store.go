package store

import (
	"encoding/json"
	"os"
	"time"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/services"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func Save(s services.MetricService, filename string) error {
	metrics, err := s.GetAllMetrics()
	if err != nil {
		logger.Log.Error("Error while getting all metrics from storage")
		return err
	}

	storeJSON, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("Error while marshal json store")
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Fatal(err.Error())
		return err
	}

	_, err = file.Write(storeJSON)
	if err != nil {
		logger.Log.Error("Error while saving file")
		return err
	}
	file.Close()

	logger.Log.Info("Success save store to file " + filename)
	return nil
}

func StartSaving(s services.MetricService, interval int, filename string) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)

		Save(s, filename)
	}
}

func Restore(s services.MetricService, filename string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		logger.Log.Error("Can't open file")
		return
	}
	var storage storage.Database
	err = json.Unmarshal(file, &storage)
	if err != nil {
		logger.Log.Error("Error while unmarshal json")
	}

	err = s.SaveFromFile(storage)
	if err != nil {
		logger.Log.Error("Error while restore from file")
	}

	logger.Log.Info("Success restore data from file " + filename)
}
