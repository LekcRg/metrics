package storage

import "context"

// Gauge — значение метрики типа gauge.
type Gauge float64

// Counter — значение метрики типа counter.
type Counter int64

// GaugeCollection — набор gauge-метрик, сгруппированных по имени.
type GaugeCollection map[string]Gauge

// CounterCollection — набор counter-метрик, сгруппированных по имени.
type CounterCollection map[string]Counter

// Database — структура, содержащая метрики типов gauge и counter.
type Database struct {
	Gauge   GaugeCollection
	Counter CounterCollection
}

// Storage — интерфейс для работы с хранилищем метрик.
// Позволяет обновлять, читать.
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
