package update

import (
	"context"

	"github.com/LekcRg/metrics/internal/models"
)

type MetricService interface {
	UpdateMetric(ctx context.Context, reqName string, reqType string, reqValue string) error
	UpdateMetricJSON(ctx context.Context, json models.Metrics) (models.Metrics, error)
	UpdateMany(ctx context.Context, list []models.Metrics) error
}
