package value

import (
	"context"

	"github.com/LekcRg/metrics/internal/models"
)

// MetricGetter — интерфейс для получения значения одной метрики.
type MetricGetter interface {
	GetMetric(ctx context.Context, reqName string, reqType string) (string, error)
	GetMetricJSON(ctx context.Context, json models.Metrics) (models.Metrics, error)
}
