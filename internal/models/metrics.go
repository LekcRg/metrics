package models

import "github.com/LekcRg/metrics/internal/server/storage"

// Metrics модель метрики типа gauge или counter.
type Metrics struct {
	Delta *storage.Counter `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *storage.Gauge   `json:"value,omitempty"` // значение метрики в случае передачи gauge
	ID    string           `json:"id"`              // имя метрики
	MType string           `json:"type"`            // параметр, принимающий значение gauge или counter
}
