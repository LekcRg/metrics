package metric

import (
	"context"
	"strconv"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func (s *MetricService) UpdateMetric(ctx context.Context, reqName string, reqType string, reqValue string) error {
	// TODO: Check errors after db
	switch reqType {
	case "counter":
		value, err := strconv.ParseInt(reqValue, 0, 64)
		if err != nil {
			return merrors.ErrIncorrectCounterValue
		}
		s.db.UpdateCounter(ctx, reqName, storage.Counter(value))
	case "gauge":
		value, err := strconv.ParseFloat(reqValue, 64)
		if err != nil {
			return merrors.ErrIncorrectGaugeValue
		}
		s.db.UpdateGauge(ctx, reqName, storage.Gauge(value))
	default:
		return merrors.ErrIncorrectMetricType
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
		return models.Metrics{}, merrors.ErrMissingMetricValue
	}
	if json.MType != "counter" {
		return models.Metrics{}, merrors.ErrIncorrectMetricType
	}

	newVal, err := s.db.UpdateCounter(ctx, json.ID, *json.Delta)

	if err != nil {
		logger.Log.Error("error while getting new counter value")
		return models.Metrics{}, merrors.ErrCannotGetNewMetricValue
	}

	return models.Metrics{
		ID:    json.ID,
		MType: json.MType,
		Delta: &newVal,
	}, nil
}

func (s *MetricService) HandleGaugeUpdate(ctx context.Context, json models.Metrics) (models.Metrics, error) {
	if json.Value == nil {
		return models.Metrics{}, merrors.ErrMissingMetricValue
	}
	if json.MType != "gauge" {
		return models.Metrics{}, merrors.ErrIncorrectMetricType
	}
	newVal, err := s.db.UpdateGauge(ctx, json.ID, *json.Value)

	if err != nil {
		logger.Log.Error("error while getting new gauge value")
		return models.Metrics{}, merrors.ErrCannotGetNewMetricValue
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
		return models.Metrics{}, merrors.ErrIncorrectMetricType
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
