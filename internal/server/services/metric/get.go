package metric

import (
	"context"
	"fmt"
	"strconv"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
)

func (s *MetricService) GetMetric(ctx context.Context, reqName string, reqType string) (string, error) {
	var (
		resVal string
		err    error
	)

	switch reqType {
	case "counter":
		var val storage.Counter
		val, err = s.db.GetCounterByName(ctx, reqName)
		resVal = fmt.Sprintf("%d", val)
	case "gauge":
		var val storage.Gauge
		val, err = s.db.GetGaugeByName(ctx, reqName)
		resVal = strconv.FormatFloat(float64(val), 'f', -1, 64)
	default:
		return "", merrors.ErrIncorrectMetricType
	}

	if err != nil {
		return "", merrors.ErrNotFoundMetric
	}

	return resVal, nil
}

func (s *MetricService) GetMetricJSON(ctx context.Context, json models.Metrics) (models.Metrics, error) {
	reqType := json.MType
	reqName := json.ID

	switch reqType {
	case "counter":
		val, err := s.db.GetCounterByName(ctx, reqName)
		if err != nil {
			logger.Log.Info("not found counter value")
			return models.Metrics{}, merrors.ErrNotFoundMetric
		}

		return models.Metrics{
			ID:    reqName,
			MType: reqType,
			Delta: &val,
		}, nil
	case "gauge":
		val, err := s.db.GetGaugeByName(ctx, reqName)
		if err != nil {
			logger.Log.Error("not found gauge value")
			return models.Metrics{}, merrors.ErrNotFoundMetric
		}

		return models.Metrics{
			ID:    reqName,
			MType: reqType,
			Value: &val,
		}, nil
	}

	return models.Metrics{}, merrors.ErrIncorrectMetricType
}

func (s *MetricService) GetAllMetrics(ctx context.Context) (storage.Database, error) {
	all, err := s.db.GetAll(ctx)
	if err != nil {
		return storage.Database{}, fmt.Errorf("something went wrong")
	}

	return all, nil
}
