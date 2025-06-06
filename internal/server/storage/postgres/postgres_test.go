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

// TODO: cycles from ../testdata
// TODO: add more tests from postgres

var (
	dbName     = "users"
	dbUser     = "user"
	dbPassword = "password"
)

func terminateContainer(t *testing.T, container testcontainers.Container) {
	require.NoError(
		t,
		testcontainers.TerminateContainer(container),
		"failed to terminate container",
	)
}

func startPostgresContainer(t *testing.T) *postgres.PostgresContainer {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		require.NoError(t, err, "failed to start container")
		return nil
	}

	return postgresContainer
}

func getPostgres(t *testing.T) (*Postgres, *postgres.PostgresContainer) {
	container := startPostgresContainer(t)
	require.NotNil(t, container)
	ctx := context.Background()

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	cfg := config.ServerConfig{
		DatabaseDSN: fmt.Sprintf(
			"postgresql://%s:%s@%s/%s?sslmode=disable",
			dbUser,
			dbPassword,
			endpoint,
			dbName,
		),
	}

	pg, err := NewPostgres(ctx, cfg)
	require.NoError(t, err)

	return pg, container
}

func TestGetUpdateGauge(t *testing.T) {
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
			ctx := context.Background()

			if !tt.wantErr {
				_, err := pg.UpdateGauge(ctx, tt.key, tt.value)
				require.NoError(t, err)
			}

			got, err := pg.GetGaugeByName(ctx, tt.key)

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
			ctx := context.Background()

			if !tt.wantErr {
				_, err := pg.UpdateCounter(ctx, tt.key, tt.value)
				require.NoError(t, err)
			}

			got, err := pg.GetCounterByName(ctx, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, got)
			}

			_, err = pg.UpdateCounter(ctx, tt.key, tt.value)
			require.NoError(t, err)
			got, err = pg.GetCounterByName(ctx, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.value*2, got)
		})
	}
}

func TestUpdateMany(t *testing.T) {
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
			if !tt.wantErr {
				err := pg.UpdateMany(ctx, tt.list)
				require.NoError(t, err)
			}

			got, err := pg.GetAll(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.list, got)

			gotCounters, err := pg.GetAllCounter(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.list.Counter, gotCounters)

			gotGauges, err := pg.GetAllGauge(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.list.Gauge, gotGauges)
		})
	}
}
