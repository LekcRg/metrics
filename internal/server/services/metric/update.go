package metric

import (
	"context"
	"errors"
	"strconv"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

var (
	ErrIncorrectCounterValue = errors.New("counter value must be int64")
	ErrIncorrectGaugeValue   = errors.New("counter value must be float64")
	ErrMissingValue          = errors.New("missing metric value")
	ErrCannotGetValue        = errors.New("can'not get new value")
)

// TODO: Check errors after db
func (s *MetricService) UpdateMetric(ctx context.Context, reqName string, reqType string, reqValue string) error {
	if reqType == "counter" {
		value, err := strconv.ParseInt(reqValue, 0, 64)
		if err != nil {
			return ErrIncorrectCounterValue
		}
		s.db.UpdateCounter(ctx, reqName, storage.Counter(value))
	} else if reqType == "gauge" {
		value, err := strconv.ParseFloat(reqValue, 64)
		if err != nil {
			return ErrIncorrectGaugeValue
		}
		s.db.UpdateGauge(ctx, reqName, storage.Gauge(value))
	} else {
		return ErrIncorrectType
	}

	if s.Config.SyncSave {
		err := s.store.Save(ctx)
		if err != nil {
			logger.Log.Error("Error while saving store")
		}
	}

	return nil
}

func (s *MetricService) HandleCounterUpdate(ctx context.Context, json models.Metrics) (models.Metrics, error) {
	if json.Delta == nil {
		return models.Metrics{}, ErrMissingValue
	}
	if json.MType != "counter" {
		return models.Metrics{}, ErrIncorrectType
	}

	newVal, err := s.db.UpdateCounter(ctx, json.ID, *json.Delta)

	if err != nil {
		logger.Log.Error("error while getting new counter value")
		return models.Metrics{}, ErrCannotGetValue
	}

	return models.Metrics{
		ID:    json.ID,
		MType: json.MType,
		Delta: &newVal,
	}, nil
}

func (s *MetricService) HandleGaugeUpdate(ctx context.Context, json models.Metrics) (models.Metrics, error) {
	if json.Value == nil {
		return models.Metrics{}, ErrMissingValue
	}
	if json.MType != "gauge" {
		return models.Metrics{}, ErrIncorrectType
	}
	newVal, err := s.db.UpdateGauge(ctx, json.ID, *json.Value)

	if err != nil {
		logger.Log.Error("error while getting new gauge value")
		return models.Metrics{}, ErrCannotGetValue
	}

	if s.Config.SyncSave {
		err := s.store.Save(ctx)
		if err != nil {
			logger.Log.Error("Error while saving store")
		}
	}

	return models.Metrics{
		ID:    json.ID,
		MType: json.MType,
		Value: &newVal,
	}, nil
}

func (s *MetricService) UpdateMetricJSON(ctx context.Context, json models.Metrics) (models.Metrics, error) {
	switch json.MType {
	case "gauge":
		return s.HandleGaugeUpdate(ctx, json)
	case "counter":
		return s.HandleCounterUpdate(ctx, json)
	default:
		return models.Metrics{}, ErrIncorrectType
	}
}

func (s *MetricService) UpdateMany(ctx context.Context, list []models.Metrics) error {
	newVals := storage.Database{
		Gauge:   storage.GaugeCollection{},
		Counter: storage.CounterCollection{},
	}

	if len(list) == 0 {
		return nil
	}

	for _, el := range list {
		if el.MType == "gauge" && el.Value != nil {
			newVals.Gauge[el.ID] = *el.Value
		} else if el.MType == "counter" && el.Delta != nil {
			newVals.Counter[el.ID] += *el.Delta
		}
	}

	return s.db.UpdateMany(ctx, newVals)
}
