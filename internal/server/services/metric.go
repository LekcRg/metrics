package services

import (
	"fmt"
	"strconv"

	"github.com/LekcRg/metrics/internal/server/storage"
)

type database interface {
	UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
	UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
	GetGaugeByName(name string) (storage.Gauge, error)
	GetCounterByName(name string) (storage.Counter, error)
	GetAll() (storage.Database, error)
}

type MetricService interface {
	UpdateMetric(reqType string, reqValue string, reqName string) error
	GetMetric(reqName string, reqType string) (string, error)
	GetAllMetrics() (storage.Database, error)
}

type metricService struct {
	db database
}

func NewMetricsService(db database) MetricService {
	return &metricService{
		db: db,
	}
}

func (s *metricService) UpdateMetric(reqName string, reqType string, reqValue string) error {
	fmt.Println(reqType, reqValue, reqName)
	if reqType == "counter" {
		value, err := strconv.ParseInt(reqValue, 0, 64)
		if err != nil {
			return fmt.Errorf("counter value must be int64")
		}
		s.db.UpdateCounter(reqName, storage.Counter(value))
	} else if reqType == "gauge" {
		value, err := strconv.ParseFloat(reqValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value must be float64")
		}
		s.db.UpdateGauge(reqName, storage.Gauge(value))
	} else {
		return fmt.Errorf("incorrect type. type must be a counter or a gauge")
	}

	return nil
}

func (s *metricService) GetMetric(reqName string, reqType string) (string, error) {
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

func (s *metricService) GetAllMetrics() (storage.Database, error) {
	all, err := s.db.GetAll()
	if err != nil {
		return storage.Database{}, fmt.Errorf("something went wrong")
	}

	return all, nil
}
