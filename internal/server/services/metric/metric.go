package metric

import (
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type Store interface {
	Save() error
}

type MetricService struct {
	Config config.ServerConfig
	db     storage.Storage
	store  *store.Store
}

func NewMetricsService(db storage.Storage, config config.ServerConfig, store *store.Store) *MetricService {
	return &MetricService{
		Config: config,
		db:     db,
		store:  store,
	}
}
