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
	Config config.ServerConfig
	db     storage.Storage
	store  Store
}

func NewMetricsService(db storage.Storage, config config.ServerConfig, store Store) *MetricService {
	return &MetricService{
		Config: config,
		db:     db,
		store:  store,
	}
}
