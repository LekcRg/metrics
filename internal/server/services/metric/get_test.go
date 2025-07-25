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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStorageArgs struct {
	wantErr      error
	st           *mocks.MockStorage
	reqName      string
	reqType      string
	gaugeDBVal   storage.Gauge
	counterDBVal storage.Counter
}

var ctx = context.Background()

func mockStorageGet(tt mockStorageArgs) {
	switch tt.reqType {
	case "counter":
		tt.st.EXPECT().GetCounterByName(ctx, tt.reqName).
			Return(tt.counterDBVal, tt.wantErr)
	case "gauge":
		tt.st.EXPECT().GetGaugeByName(ctx, tt.reqName).
			Return(tt.gaugeDBVal, tt.wantErr)
	}
}

const (
	counterName = "counter-name"
	gaugeName   = "gauge-name"
	notFound    = "not-found"
)

var (
	counterVal = storage.Counter(123)
	gaugeVal   = storage.Gauge(1.23)
)

func TestGetMetric(t *testing.T) {
	type args struct {
		reqName string
		reqType string
	}
	tests := []struct {
		wantErr      error
		args         args
		name         string
		want         string
		counterDBVal storage.Counter
		gaugeDBVal   storage.Gauge
	}{
		{
			name: "Get default counter",
			args: args{
				reqName: counterName,
				reqType: "counter",
			},
			counterDBVal: counterVal,
			want:         strconv.Itoa(int(counterVal)),
			wantErr:      nil,
		},
		{
			name: "Get default gauge",
			args: args{
				reqName: gaugeName,
				reqType: "gauge",
			},
			gaugeDBVal: gaugeVal,
			want:       strconv.FormatFloat(float64(gaugeVal), 'f', -1, 64),
			wantErr:    nil,
		},
		{
			name: "Get not found counter",
			args: args{
				reqName: notFound,
				reqType: "counter",
			},
			want:    "",
			wantErr: merrors.ErrNotFoundMetric,
		},
		{
			name: "Get not found gauge",
			args: args{
				reqName: notFound,
				reqType: "gauge",
			},
			want:    "",
			wantErr: merrors.ErrNotFoundMetric,
		},
		{
			name: "Get incorrect type",
			args: args{
				reqName: "incorrect",
				reqType: "test",
			},
			want:    "",
			wantErr: merrors.ErrIncorrectMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)

			mockStorageGet(mockStorageArgs{
				st:           st,
				reqName:      tt.args.reqName,
				reqType:      tt.args.reqType,
				wantErr:      tt.wantErr,
				gaugeDBVal:   tt.gaugeDBVal,
				counterDBVal: tt.counterDBVal,
			})

			s := &MetricService{
				Config: config.ServerConfig{},
				db:     st,
				store:  NewMockStore(t),
			}
			got, err := s.GetMetric(ctx, tt.args.reqName, tt.args.reqType)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMetricJSON(t *testing.T) {
	counterModel := models.Metrics{
		ID:    counterName,
		MType: "counter",
		Delta: &counterVal,
	}
	gaugeModel := models.Metrics{
		ID:    gaugeName,
		MType: "gauge",
		Value: &gaugeVal,
	}
	tests := []struct {
		wantErr      error
		arg          models.Metrics
		want         models.Metrics
		name         string
		counterDBVal storage.Counter
		gaugeDBVal   storage.Gauge
	}{
		{
			name:         "Get default counter",
			arg:          counterModel,
			want:         counterModel,
			wantErr:      nil,
			counterDBVal: counterVal,
		},
		{
			name:       "Get default gauge",
			arg:        gaugeModel,
			want:       gaugeModel,
			wantErr:    nil,
			gaugeDBVal: gaugeVal,
		},
		{
			name: "Get not found counter",
			arg: models.Metrics{
				ID:    notFound,
				MType: "counter",
			},
			want:    models.Metrics{},
			wantErr: merrors.ErrNotFoundMetric,
		},
		{
			name: "Get not found gauge",
			arg: models.Metrics{
				ID:    notFound,
				MType: "gauge",
			},
			want:    models.Metrics{},
			wantErr: merrors.ErrNotFoundMetric,
		},
		{
			name: "Get incorrect type",
			arg: models.Metrics{
				ID:    "incorrect",
				MType: "test",
			},
			want:    models.Metrics{},
			wantErr: merrors.ErrIncorrectMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := mocks.NewMockStorage(t)

			mockStorageGet(mockStorageArgs{
				st:           st,
				reqName:      tt.arg.ID,
				reqType:      tt.arg.MType,
				wantErr:      tt.wantErr,
				gaugeDBVal:   tt.gaugeDBVal,
				counterDBVal: tt.counterDBVal,
			})

			s := &MetricService{
				Config: config.ServerConfig{},
				db:     st,
				store:  NewMockStore(t),
			}
			got, err := s.GetMetricJSON(ctx, tt.arg)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			// TODO: equal struct, maybe errors
			assert.Equal(t, tt.want, got)
			switch tt.arg.MType {
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
