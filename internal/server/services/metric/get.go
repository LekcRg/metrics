package metric

import (
	"fmt"
	"strconv"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func (s *MetricService) GetMetric(reqName string, reqType string) (string, error) {
	var (
		resVal string
		err    error
	)

	if reqType == "counter" {
		var val storage.Counter
		val, err = s.db.GetCounterByName(reqName)
		resVal = fmt.Sprintf("%d", val)
	} else if reqType == "gauge" {
		var val storage.Gauge
		val, err = s.db.GetGaugeByName(reqName)
		resVal = strconv.FormatFloat(float64(val), 'f', -1, 64)
	}

	if err != nil {
		return "", fmt.Errorf("not found")
	}

	return resVal, nil
}

func (s *MetricService) GetMetricJSON(json models.Metrics) (models.Metrics, error) {
	reqType := json.MType
	reqName := json.ID

	if reqType == "counter" {
		val, err := s.db.GetCounterByName(reqName)
		if err != nil {
			logger.Log.Error("error while getting counter value")
			return models.Metrics{}, fmt.Errorf("can'not get new value")
		}

		return models.Metrics{
			ID:    reqName,
			MType: reqType,
			Delta: &val,
		}, nil
	} else if reqType == "gauge" {
		val, err := s.db.GetGaugeByName(reqName)
		if err != nil {
			logger.Log.Error("error while getting gauge value")
			return models.Metrics{}, fmt.Errorf("can'not get new value")
		}

		return models.Metrics{
			ID:    reqName,
			MType: reqType,
			Value: &val,
		}, nil
	}

	return models.Metrics{}, fmt.Errorf("type is not valid")
}

func (s *MetricService) GetAllMetrics() (storage.Database, error) {
	all, err := s.db.GetAll()
	if err != nil {
		return storage.Database{}, fmt.Errorf("something went wrong")
	}

	return all, nil
}
