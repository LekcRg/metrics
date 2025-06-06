package testdata

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func CycleGetUpdateGauge(t *testing.T, func) {
	pg, container := getPostgres(t)
	defer terminateContainer(t, container)

	tests := []struct {
		name    string
		key     string
		value   storage.Gauge
		wantErr bool
	}{
		{
			name:    "Set 42.42",
			key:     "gauge1",
			value:   42.42,
			wantErr: false,
		},
		{
			name:    "Set 0.0",
			key:     "gauge2",
			value:   0.0,
			wantErr: false,
		},
		{
			name:    "Read unknown key",
			key:     "missing",
			value:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: func()
		})
	}
}

func CycleUpdateCounter(t *testing.T) {
	pg, container := getPostgres(t)
	defer terminateContainer(t, container)

	tests := []struct {
		name    string
		key     string
		value   storage.Counter
		wantErr bool
		twice   bool
	}{
		{
			name:    "Set 42",
			key:     "counter1",
			value:   42,
			wantErr: false,
		},
		{
			name:    "Set 0",
			key:     "counter2",
			value:   0,
			wantErr: false,
		},
		{
			name:    "Set 1 twice",
			key:     "counter2",
			value:   1,
			wantErr: false,
		},
		{
			name:    "Read unknown key",
			key:     "missing",
			value:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: func()
		})
	}
}

func CycleUpdateMany(t *testing.T) {
	ctx := context.Background()
	pg, container := getPostgres(t)
	defer terminateContainer(t, container)

	counters := storage.CounterCollection{
		"counter1": 42,
		"counter2": 0,
	}
	gauges := storage.GaugeCollection{
		"gauge1": 42.42,
		"gauge2": 0,
	}
	db := storage.Database{
		Counter: counters,
		Gauge:   gauges,
	}

	tests := []struct {
		list    storage.Database
		name    string
		wantErr bool
	}{
		{
			name:    "Positive",
			list:    db,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: func()
		})
	}
}
