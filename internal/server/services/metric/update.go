package metric

import (
	"context"
	"fmt"
	"strconv"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func (s *MetricService) UpdateMetric(ctx context.Context, reqName string, reqType string, reqValue string) error {
	if reqType == "counter" {
		value, err := strconv.ParseInt(reqValue, 0, 64)
		if err != nil {
			return fmt.Errorf("counter value must be int64")
		}
		s.db.UpdateCounter(ctx, reqName, storage.Counter(value))
	} else if reqType == "gauge" {
		value, err := strconv.ParseFloat(reqValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value must be float64")
		}
		s.db.UpdateGauge(ctx, reqName, storage.Gauge(value))
	} else {
		return fmt.Errorf("incorrect type. type must be a counter or a gauge")
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
	newVal, err := s.db.UpdateCounter(ctx, json.ID, storage.Counter(*json.Delta))

	if err != nil {
		logger.Log.Error("error while getting new counter value")
		return models.Metrics{}, fmt.Errorf("can'not get new value")
	}

	return models.Metrics{
		ID:    json.ID,
		MType: json.MType,
		Delta: &newVal,
	}, nil
}

func (s *MetricService) HandleGaugeUpdate(ctx context.Context, json models.Metrics) (models.Metrics, error) {
	newVal, err := s.db.UpdateGauge(ctx, json.ID, storage.Gauge(*json.Value))

	if err != nil {
		logger.Log.Error("error while getting new gauge value")
		return models.Metrics{}, fmt.Errorf("can'not get new value")
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
	switch {
	case json.MType == "gauge" && json.Value != nil:
		return s.HandleGaugeUpdate(ctx, json)
	case json.MType == "counter" && json.Delta != nil:
		return s.HandleCounterUpdate(ctx, json)
	default:
		return models.Metrics{}, fmt.Errorf("invalid type or empty value")
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
