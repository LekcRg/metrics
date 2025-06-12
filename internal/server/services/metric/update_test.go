package metric

import (
	"context"
	"strconv"
	"testing"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/LekcRg/metrics/internal/mocks"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/LekcRg/metrics/internal/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateArgs struct {
	wantErr  error
	st       *mocks.MockStorage
	reqName  string
	reqValue string
	reqType  string
}

func TestUpdateMetric(t *testing.T) {
	type args struct {
		reqName  string
		reqType  string
		reqValue string
	}
	tests := []struct {
		wantErr error
		args    args
		name    string
		save    bool
		saveErr bool
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
			wantErr: merrors.ErrIncorrectMetricType,
		},
		{
			name: "Create gauge with incorrect value",
			args: args{
				reqName:  gaugeName,
				reqType:  "gauge",
				reqValue: "incorrect",
			},
			wantErr: merrors.ErrIncorrectGaugeValue,
		},
		{
			name: "Create counter with incorrect value",
			args: args{
				reqName:  counterName,
				reqType:  "counter",
				reqValue: "incorrect",
			},
			wantErr: merrors.ErrIncorrectCounterValue,
		},
		{
			name: "Sync save",
			args: args{
				reqName:  "counterName",
				reqType:  "counter",
				reqValue: "10",
			},
			wantErr: nil,
			save:    true,
		},
		{
			name: "Sync save error",
			args: args{
				reqName:  "counterName",
				reqType:  "counter",
				reqValue: "10",
			},
			wantErr: nil,
			save:    true,
			saveErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)
			ctx := context.Background()

			switch tt.args.reqType {
			case "counter":
				valInt, err := strconv.ParseInt(tt.args.reqValue, 0, 64)
				// TODO: check err?
				if err == nil {
					val := storage.Counter(valInt)
					st.EXPECT().UpdateCounter(ctx, tt.args.reqName, val).Return(val, tt.wantErr)
				}
			case "gauge":
				valFloat, err := strconv.ParseFloat(tt.args.reqValue, 64)
				// TODO: check err?
				if err == nil {
					val := storage.Gauge(valFloat)
					st.EXPECT().UpdateGauge(ctx, tt.args.reqName, val).Return(val, tt.wantErr)
				}
			}

			store := NewMockStore(t)
			if tt.save {
				var err error

				if tt.saveErr {
					err = merrors.ErrMocked
				}
				store.EXPECT().Save(context.Background()).Return(err)
			}

			s := &MetricService{
				Config: config.ServerConfig{
					SyncSave: tt.save,
				},
				db:    st,
				store: store,
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
		wantErr error
		json    models.Metrics
		want    models.Metrics
		name    string
		dbErr   bool
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
			wantErr: merrors.ErrIncorrectMetricType,
		},
		{
			name: "Counter with nil value",
			json: models.Metrics{
				ID:    counterName,
				MType: "counter",
			},
			want:    models.Metrics{},
			wantErr: merrors.ErrMissingMetricValue,
		},
		{
			name:    "DB return error",
			json:    validJSON,
			want:    validJSON,
			wantErr: merrors.ErrCannotGetNewMetricValue,
			dbErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.wantErr == nil || tt.dbErr {
				var err error = nil
				if tt.dbErr {
					err = merrors.ErrMocked
				}
				st.EXPECT().UpdateCounter(ctx, tt.json.ID, *tt.json.Delta).
					Return(*tt.want.Delta, err)
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  NewMockStore(t),
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
		wantErr error
		json    models.Metrics
		want    models.Metrics
		name    string
		save    bool
		saveErr bool
		dbErr   bool
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
			wantErr: merrors.ErrIncorrectMetricType,
		},
		{
			name: "Gauge with nil value",
			json: models.Metrics{
				ID:    gaugeName,
				MType: "gauge",
			},
			want:    models.Metrics{},
			wantErr: merrors.ErrMissingMetricValue,
		},
		{
			name:    "DB return error",
			json:    validJSON,
			want:    validJSON,
			wantErr: merrors.ErrCannotGetNewMetricValue,
			dbErr:   true,
		},
		{
			name: "Sync save",
			json: validJSON,
			want: validJSON,
			save: true,
		},
		{
			name:    "Sync save error",
			json:    validJSON,
			want:    validJSON,
			save:    true,
			saveErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.wantErr == nil || tt.dbErr {
				var err error

				if tt.dbErr {
					err = merrors.ErrMocked
				}
				st.EXPECT().UpdateGauge(ctx, tt.json.ID, *tt.json.Value).
					Return(*tt.want.Value, err)
			}

			store := NewMockStore(t)
			if tt.save {
				var err error

				if tt.saveErr {
					err = merrors.ErrMocked
				}
				store.EXPECT().Save(context.Background()).Return(err)
			}

			s := &MetricService{
				Config: config.ServerConfig{
					SyncSave: tt.save,
				},
				db:    st,
				store: store,
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
		wantErr error
		json    models.Metrics
		want    models.Metrics
		name    string
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
			wantErr: merrors.ErrIncorrectMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)
			ctx := context.Background()

			if tt.wantErr == nil {
				switch tt.json.MType {
				case "counter":
					st.EXPECT().UpdateCounter(ctx, tt.json.ID, *tt.json.Delta).
						Return(*tt.want.Delta, tt.wantErr)
				case "gauge":
					st.EXPECT().UpdateGauge(ctx, tt.json.ID, *tt.json.Value).
						Return(*tt.want.Value, tt.wantErr)
				}
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  NewMockStore(t),
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

			switch tt.json.MType {
			case "counter":
				require.NotNil(t, got.Delta)
				assert.Equal(t, *tt.want.Delta, *got.Delta)
			case "gauge":
				require.NotNil(t, got.Value)
				assert.Equal(t, *tt.want.Value, *got.Value)
			}
		})
	}
}

func TestUpdateMany(t *testing.T) {
	tests := []struct {
		wantDBData storage.Database
		dbErr      error
		wantErr    error
		name       string
		metrics    []models.Metrics
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
			dbErr:   merrors.ErrMocked,
			wantErr: merrors.ErrMocked,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)
			ctx := context.Background()

			// Ожидание только если что-то должно быть
			if len(tt.metrics) > 0 {
				st.EXPECT().UpdateMany(ctx, tt.wantDBData).Return(tt.dbErr)
			}

			s := &MetricService{
				Config: testdata.TestServerConfig,
				db:     st,
				store:  NewMockStore(t),
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
