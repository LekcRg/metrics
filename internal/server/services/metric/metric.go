package metric

import (
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type Store interface {
	Save() error
}

type Storage interface {
	UpdateCounter(name string, value storage.Counter) (storage.Counter, error)
	UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error)
	GetGaugeByName(name string) (storage.Gauge, error)
	GetCounterByName(name string) (storage.Counter, error)
	GetAll() (storage.Database, error)
	SaveManyGauge(storage.GaugeCollection) error
	SaveManyCounter(storage.CounterCollection) error
}

type MetricService struct {
	config config.ServerConfig
	db     Storage
	store  *store.Store
}

func NewMetricsService(db Storage, config config.ServerConfig, store *store.Store) *MetricService {
	return &MetricService{
		config: config,
		db:     db,
		store:  store,
	}
}
