package metric

import (
	"context"
	"errors"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/storage"
)

var (
	ErrIncorrectType = errors.New("incorrect type. type must be a counter or a gauge")
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
