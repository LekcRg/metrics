package value

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPost(t *testing.T) {
	counter := &models.Metrics{
		ID:    "one",
		MType: "counter",
		Value: ptrGauge(12.34),
	}
	gauge := &models.Metrics{
		ID:    "one",
		MType: "gauge",
		Delta: ptrCounter(12),
	}
	type want struct {
		code   int
		metric *models.Metrics
	}
	tests := []struct {
		name         string
		serviceError bool
		body         string
		input        *models.Metrics
		want         want
	}{
		{
			name:  "Valid gauge",
			input: gauge,
			want: want{
				code:   http.StatusOK,
				metric: gauge,
			},
		},
		{
			name:  "Valid counter",
			input: counter,
			want: want{
				code:   http.StatusOK,
				metric: counter,
			},
		},
		{
			name: "Invalid JSON",
			body: "{invalid}",
			want: want{
				code:   http.StatusInternalServerError,
				metric: counter,
			},
		},
		{
			name:         "Service return error",
			serviceError: true,
			input:        counter,
			want: want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockMetricGetter(t)

			w := httptest.NewRecorder()

			var body io.Reader
			if tt.input != nil {
				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(tt.input)
				require.NoError(t, err)
				body = buf
			} else {
				body = strings.NewReader(tt.body)
			}

			r := httptest.NewRequest(http.MethodPost, "/value", body)

			if tt.input != nil || tt.serviceError {
				var err error = nil

				if tt.serviceError {
					err = merrors.ErrMocked
				}
				var wantMetric models.Metrics
				if tt.want.metric != nil {
					wantMetric = *tt.want.metric
				}
				s.EXPECT().
					GetMetricJSON(context.Background(), *tt.input).
					Return(wantMetric, err)
			}

			contentType := "application/json"
			r.Header.Set("Content-Type", contentType)

			h := Post(s)
			h(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}

func ptrGauge(val float64) *storage.Gauge {
	res := storage.Gauge(val)

	return &res
}

func ptrCounter(val int64) *storage.Counter {
	res := storage.Counter(val)

	return &res
}
