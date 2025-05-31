package memstorage

import (
	"context"

	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type MemStorage struct {
	db *storage.Database
}

func New() (*MemStorage, error) {
	return &MemStorage{
		db: &storage.Database{
			Gauge:   make(storage.GaugeCollection),
			Counter: make(storage.CounterCollection),
		},
	}, nil
}

func (s *MemStorage) UpdateCounter(_ context.Context, name string, value storage.Counter) (storage.Counter, error) {
	s.db.Counter[name] += value

	return s.db.Counter[name], nil
}

func (s *MemStorage) UpdateGauge(_ context.Context, name string, value storage.Gauge) (storage.Gauge, error) {
	s.db.Gauge[name] = value

	return s.db.Gauge[name], nil
}

func (s *MemStorage) UpdateMany(_ context.Context, list storage.Database) error {
	for key, item := range list.Gauge {
		s.db.Gauge[key] = item
	}

	for key, item := range list.Counter {
		s.db.Counter[key] += item
	}

	return nil
}

func (s *MemStorage) GetAllCounter(_ context.Context) (storage.CounterCollection, error) {
	collection := s.db.Counter
	return collection, nil
}

func (s *MemStorage) GetAllGauge(_ context.Context) (storage.GaugeCollection, error) {
	collection := s.db.Gauge
	return collection, nil
}

func (s *MemStorage) GetGaugeByName(_ context.Context, name string) (storage.Gauge, error) {
	if val, ok := s.db.Gauge[name]; ok {
		return val, nil
	}

	return 0, merrors.ErrNotFoundMetric
}

func (s *MemStorage) GetCounterByName(_ context.Context, name string) (storage.Counter, error) {
	if val, ok := s.db.Counter[name]; ok {
		return val, nil
	}

	return 0, merrors.ErrNotFoundMetric
}

func (s *MemStorage) GetAll(_ context.Context) (storage.Database, error) {
	return *s.db, nil
}

func (s *MemStorage) Ping(_ context.Context) error {
	return nil
}

func (s MemStorage) Close() {
	//
}
