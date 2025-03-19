package storage

import "context"

type Gauge float64
type Counter int64
type GaugeCollection map[string]Gauge
type CounterCollection map[string]Counter
type Database struct {
	Gauge   GaugeCollection
	Counter CounterCollection
}
type Storage interface {
	UpdateCounter(ctx context.Context, name string, value Counter) (Counter, error)
	UpdateGauge(ctx context.Context, name string, value Gauge) (Gauge, error)
	UpdateMany(ctx context.Context, list Database) error
	GetGaugeByName(ctx context.Context, name string) (Gauge, error)
	GetCounterByName(ctx context.Context, name string) (Counter, error)
	GetAll(ctx context.Context) (Database, error)
	Ping(ctx context.Context) error
	Close()
}
