package storage

type Gauge float64
type Counter int64
type GaugeCollection map[string]Gauge
type CounterCollection map[string]Counter
type Database struct {
	Gauge   GaugeCollection
	Counter CounterCollection
}
type Storage interface {
	UpdateCounter(name string, value Counter) (Counter, error)
	UpdateGauge(name string, value Gauge) (Gauge, error)
	GetGaugeByName(name string) (Gauge, error)
	GetCounterByName(name string) (Counter, error)
	GetAll() (Database, error)
	SaveManyGauge(GaugeCollection) error
	SaveManyCounter(CounterCollection) error
	Ping() error
}
