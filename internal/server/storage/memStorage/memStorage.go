package memStorage

import (
	"errors"
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

func (s *MemStorage) GetGaugeByName(name string) (storage.Gauge, error) {
	if val, ok := s.db.Gauge[name]; ok {
		return val, nil
	} else {
		return 0, errors.New("not found")
	}
}

func (s *MemStorage) GetCounterByName(name string) (storage.Counter, error) {
	if val, ok := s.db.Counter[name]; ok {
		return val, nil
	} else {
		return 0, errors.New("not found")
	}
}

func (s *MemStorage) GetAll() (storage.Database, error) {
	return *s.db, nil
}
