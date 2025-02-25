package services

import (
	"fmt"
	"strconv"

	"github.com/LekcRg/metrics/internal/models"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type database interface {
	UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
	UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
	GetGaugeByName(name string) (storage.Gauge, error)
	GetCounterByName(name string) (storage.Counter, error)
	GetAll() (storage.Database, error)
	SaveManyGauge(storage.GaugeCollection) error
	SaveManyCounter(storage.CounterCollection) error
}

type MetricService interface {
	UpdateMetricJSON(json models.Metrics) (models.Metrics, error)
	UpdateMetric(reqType string, reqValue string, reqName string) error
	GetMetric(reqName string, reqType string) (string, error)
	GetMetricJSON(json models.Metrics) (models.Metrics, error)
	GetAllMetrics() (storage.Database, error)
	SaveFromFile(storage.Database) error
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

func (s *metricService) UpdateMetricJSON(json models.Metrics) (models.Metrics, error) {
	reqName := json.ID
	reqType := json.MType
	delta := json.Delta
	value := json.Value
	// var newDelta *storage.Counter
	// var newValue *storage.Gauge

	if reqType == "gauge" && value != nil {
		s.db.UpdateGauge(reqName, storage.Gauge(*value))

		newVal, err := s.db.GetGaugeByName(reqName)
		if err != nil {
			logger.Log.Error("error while getting new gauge value")
			return models.Metrics{}, fmt.Errorf("can'not get new value")
		}

		return models.Metrics{
			ID:    reqName,
			MType: reqType,
			Value: &newVal,
		}, nil

	} else if reqType == "counter" && delta != nil {
		s.db.UpdateCounter(reqName, storage.Counter(*delta))

		newVal, err := s.db.GetCounterByName(reqName)
		if err != nil {
			logger.Log.Error("error while getting new counter value")
			return models.Metrics{}, fmt.Errorf("can'not get new value")
		}

		return models.Metrics{
			ID:    reqName,
			MType: reqType,
			Delta: &newVal,
		}, nil
	}
	return models.Metrics{}, fmt.Errorf("invalid type or empty value")
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

func (s *metricService) GetMetricJSON(json models.Metrics) (models.Metrics, error) {
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

func (s *metricService) GetAllMetrics() (storage.Database, error) {
	all, err := s.db.GetAll()
	if err != nil {
		return storage.Database{}, fmt.Errorf("something went wrong")
	}

	return all, nil
}

func (s *metricService) SaveFromFile(file storage.Database) error {
	if len(file.Gauge) > 0 {
		err := s.db.SaveManyGauge(file.Gauge)
		if err != nil {
			return err
		}
	}

	if len(file.Counter) > 0 {
		err := s.db.SaveManyCounter(file.Counter)
		if err != nil {
			return err
		}
	}

	return nil
}
