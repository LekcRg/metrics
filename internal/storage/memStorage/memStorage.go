package memStorage

import "github.com/LekcRg/metrics/internal/storage"

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

func (s *MemStorage) UpdateCounter(name string, value storage.Counter) (storage.Counter, error) {
	s.db.Counter[name] += value

	return s.db.Counter[name], nil
}

func (s *MemStorage) UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error) {
	s.db.Gauge[name] = value

	return s.db.Gauge[name], nil
}

func (s *MemStorage) GetAllCounter() (storage.CounterCollection, error) {
	collection := s.db.Counter
	return collection, nil
}

func (s *MemStorage) GetAllGouge() (storage.GaugeCollection, error) {
	collection := s.db.Gauge
	return collection, nil
}

func (s *MemStorage) GetAll() (storage.Database, error) {
	return *s.db, nil
}
