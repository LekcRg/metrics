package storage

type Gauge float64
type Counter int64
type GaugeCollection map[string]Gauge
type CounterCollection map[string]Counter
type Database struct {
	Gauge   GaugeCollection
	Counter CounterCollection
}
