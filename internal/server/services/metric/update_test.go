package metric

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/services/metric/mocks"
	"github.com/LekcRg/metrics/internal/server/storage"
	storagemocks "github.com/LekcRg/metrics/internal/server/storage/mocks"
	"github.com/LekcRg/metrics/internal/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateArgs struct {
	st       *storagemocks.MockStorage
	reqName  string
	reqValue string
	reqType  string
	wantErr  error
}

func TestUpdateMetric(t *testing.T) {
	type args struct {
		reqName  string
		reqType  string
		reqValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Create counter metric",
			args: args{
				reqName:  "counterName",
				reqType:  "counter",
				reqValue: "10",
			},
			wantErr: nil,
		},
		{
			name: "Create gauge metric",
			args: args{
				reqName:  gaugeName,
				reqType:  "gauge",
				reqValue: "1.23",
			},
			wantErr: nil,
		},
		{
			name: "Create metric with incorrect type",
			args: args{
				reqName:  "incorrect",
				reqType:  "incorrect",
				reqValue: "1",
			},
			wantErr: ErrIncorrectType,
		},
		{
			name: "Create gauge with incorrect value",
			args: args{
				reqName:  gaugeName,
				reqType:  "gauge",
				reqValue: "incorrect",
			},
			wantErr: ErrIncorrectGaugeValue,
		},
		{
			name: "Create counter with incorrect value",
			args: args{
				reqName:  counterName,
				reqType:  "counter",
				reqValue: "incorrect",
			},
			wantErr: ErrIncorrectCounterValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storagemocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.args.reqType == "counter" {
				valInt, err := strconv.ParseInt(tt.args.reqValue, 0, 64)
				// TODO: check err?
				if err == nil {
					val := storage.Counter(valInt)
					st.EXPECT().UpdateCounter(ctx, tt.args.reqName, val).Return(val, tt.wantErr)
				}
			} else if tt.args.reqType == "gauge" {
				valFloat, err := strconv.ParseFloat(tt.args.reqValue, 64)
				// TODO: check err?
				if err == nil {
					val := storage.Gauge(valFloat)
					st.EXPECT().UpdateGauge(ctx, tt.args.reqName, val).Return(val, tt.wantErr)
				}
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  mocks.NewMockStore(t),
			}
			err := s.UpdateMetric(ctx, tt.args.reqName, tt.args.reqType, tt.args.reqValue)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestHandleCounterUpdate(t *testing.T) {
	validJSON := models.Metrics{
		ID:    counterName,
		MType: "counter",
		Delta: &counterVal,
	}

	tests := []struct {
		name    string
		json    models.Metrics
		want    models.Metrics
		wantErr error
	}{
		{
			name:    "Valid counter",
			json:    validJSON,
			want:    validJSON,
			wantErr: nil,
		},
		{
			name: "Invalid type",
			json: models.Metrics{
				ID:    counterName,
				MType: "gauge",
				Delta: &counterVal,
			},
			want:    models.Metrics{},
			wantErr: ErrIncorrectType,
		},
		{
			name: "Counter with nil value",
			json: models.Metrics{
				ID:    counterName,
				MType: "counter",
			},
			want:    models.Metrics{},
			wantErr: ErrMissingValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storagemocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.wantErr == nil {
				st.EXPECT().UpdateCounter(ctx, tt.json.ID, *tt.json.Delta).
					Return(*tt.want.Delta, tt.wantErr)
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  mocks.NewMockStore(t),
			}
			got, err := s.HandleCounterUpdate(ctx, tt.json)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, tt.wantErr, err)

				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)

			require.NotNil(t, got.Delta)
			assert.Equal(t, *tt.want.Delta, *got.Delta)
		})
	}
}

func TestHandleGaugeUpdate(t *testing.T) {
	validJSON := models.Metrics{
		ID:    gaugeName,
		MType: "gauge",
		Value: &gaugeVal,
	}

	tests := []struct {
		name    string
		json    models.Metrics
		want    models.Metrics
		wantErr error
	}{
		{
			name:    "Valid gauge",
			json:    validJSON,
			want:    validJSON,
			wantErr: nil,
		},
		{
			name: "Invalid type",
			json: models.Metrics{
				ID:    gaugeName,
				MType: "counter",
				Value: &gaugeVal,
			},
			want:    models.Metrics{},
			wantErr: ErrIncorrectType,
		},
		{
			name: "Gauge with nil value",
			json: models.Metrics{
				ID:    gaugeName,
				MType: "gauge",
			},
			want:    models.Metrics{},
			wantErr: ErrMissingValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storagemocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.wantErr == nil {
				st.EXPECT().UpdateGauge(ctx, tt.json.ID, *tt.json.Value).
					Return(*tt.want.Value, tt.wantErr)
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  mocks.NewMockStore(t),
			}
			got, err := s.HandleGaugeUpdate(ctx, tt.json)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, tt.wantErr, err)

				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)

			require.NotNil(t, got.Value)
			assert.Equal(t, *tt.want.Value, *got.Value)
		})
	}
}

func TestUpdateMetricJSON(t *testing.T) {
	gaugeModel := models.Metrics{
		ID:    gaugeName,
		MType: "gauge",
		Value: &gaugeVal,
	}
	counterModel := models.Metrics{
		ID:    counterName,
		MType: "counter",
		Delta: &counterVal,
	}
	type args struct {
		ctx  context.Context
		json models.Metrics
	}
	tests := []struct {
		name    string
		json    models.Metrics
		want    models.Metrics
		wantErr error
	}{
		{
			name:    "Update gauge",
			json:    gaugeModel,
			want:    gaugeModel,
			wantErr: nil,
		},
		{
			name:    "Update counter",
			json:    counterModel,
			want:    counterModel,
			wantErr: nil,
		},
		{
			name: "Update invalid type",
			json: models.Metrics{
				ID:    "incorrect",
				MType: "incorrect",
				Delta: &counterVal,
			},
			want:    models.Metrics{},
			wantErr: ErrIncorrectType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storagemocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.wantErr == nil {
				if tt.json.MType == "gauge" {
					st.EXPECT().UpdateGauge(ctx, tt.json.ID, *tt.json.Value).
						Return(*tt.want.Value, tt.wantErr)
				} else if tt.json.MType == "counter" {
					st.EXPECT().UpdateCounter(ctx, tt.json.ID, *tt.json.Delta).
						Return(*tt.want.Delta, tt.wantErr)
				}
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  mocks.NewMockStore(t),
			}

			got, err := s.UpdateMetricJSON(ctx, tt.json)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, tt.wantErr, err)

				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)

			if tt.json.MType == "gauge" {
				require.NotNil(t, got.Value)
				assert.Equal(t, *tt.want.Value, *got.Value)
			} else if tt.json.MType == "counter" {
				require.NotNil(t, got.Delta)
				assert.Equal(t, *tt.want.Delta, *got.Delta)
			}
		})
	}
}

func TestUpdateMany(t *testing.T) {
	dbErr := errors.New("db error")

	tests := []struct {
		name       string
		metrics    []models.Metrics
		wantDBData storage.Database
		dbErr      error
		wantErr    error
	}{
		{
			name:       "Empty list",
			metrics:    nil,
			wantDBData: storage.Database{},
			dbErr:      nil,
			wantErr:    nil,
		},
		{
			name: "Gauge and Counter present",
			metrics: []models.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: ptrGauge(1.5),
				},
				{
					ID:    "counter1",
					MType: "counter",
					Delta: ptrCounter(3),
				},
				{
					ID:    "counter1",
					MType: "counter",
					Delta: ptrCounter(2),
				},
			},
			wantDBData: storage.Database{
				Gauge: storage.GaugeCollection{
					"gauge1": 1.5,
				},
				Counter: storage.CounterCollection{
					"counter1": 5,
				},
			},
			dbErr:   nil,
			wantErr: nil,
		},
		{
			name: "DB returns error",
			metrics: []models.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: ptrGauge(2.2),
				},
			},
			wantDBData: storage.Database{
				Gauge: storage.GaugeCollection{
					"gauge1": 2.2,
				},
				Counter: storage.CounterCollection{},
			},
			dbErr:   dbErr,
			wantErr: dbErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storagemocks.NewMockStorage(t)
			ctx := context.Background()

			// Ожидание только если что-то должно быть
			if len(tt.metrics) > 0 {
				st.EXPECT().UpdateMany(ctx, tt.wantDBData).Return(tt.dbErr)
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  mocks.NewMockStore(t),
			}

			err := s.UpdateMany(ctx, tt.metrics)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func ptrGauge(val float64) *storage.Gauge {
	v := storage.Gauge(val)
	return &v
}

func ptrCounter(val int64) *storage.Counter {
	v := storage.Counter(val)
	return &v
}
