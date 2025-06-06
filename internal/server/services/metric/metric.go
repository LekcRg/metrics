package metric

import (
	"context"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type Store interface {
	Save(ctx context.Context) error
}

type MetricService struct {
	db     storage.Storage
	store  Store
	Config config.ServerConfig
}

func NewMetricsService(db storage.Storage, config config.ServerConfig, store Store) *MetricService {
	return &MetricService{
		Config: config,
		db:     db,
		store:  store,
	}
}
