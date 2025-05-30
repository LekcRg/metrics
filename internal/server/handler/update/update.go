package update

import (
	"context"

	"github.com/LekcRg/metrics/internal/models"
)

// MetricUpdater — интерфейс для обновления одной или нескольких метрик.
type MetricUpdater interface {
	UpdateMetric(ctx context.Context, reqName string, reqType string, reqValue string) error
	UpdateMetricJSON(ctx context.Context, json models.Metrics) (models.Metrics, error)
	UpdateMany(ctx context.Context, list []models.Metrics) error
}
