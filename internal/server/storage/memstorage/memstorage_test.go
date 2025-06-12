package memstorage

import (
	"context"
	"testing"

	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: cycles from ../testdata

func TestUpdateGauge(t *testing.T) {
	ctx := context.Background()
	s, err := New()
	require.NoError(t, err)

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
			if !tt.wantErr {
				_, err := s.UpdateGauge(ctx, tt.key, tt.value)
				require.NoError(t, err)
			}

			got, err := s.GetGaugeByName(ctx, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, got)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	ctx := context.Background()
	s, err := New()
	require.NoError(t, err)

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
			if !tt.wantErr {
				_, err := s.UpdateCounter(ctx, tt.key, tt.value)
				require.NoError(t, err)
			}

			got, err := s.GetCounterByName(ctx, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, got)
			}

			if tt.twice {
				return
			}

			_, err = s.UpdateCounter(ctx, tt.key, tt.value)
			require.NoError(t, err)
			got, err = s.GetCounterByName(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.value*2, got)
		})
	}
}

func TestUpdateMany(t *testing.T) {
	ctx := context.Background()
	s, err := New()
	require.NoError(t, err)
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
			if !tt.wantErr {
				err := s.UpdateMany(ctx, tt.list)
				require.NoError(t, err)
			}

			got, err := s.GetAll(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.list, got)

			gotCounters, err := s.GetAllCounter(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.list.Counter, gotCounters)

			gotGauges, err := s.GetAllGauge(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.list.Gauge, gotGauges)
		})
	}
}
