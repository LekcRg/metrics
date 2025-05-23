package update

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/handler/update/mocks"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(v)
	require.NoError(t, err)
	return b
}

func TestPostJSON(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name         string
		input        *models.Metrics
		contentType  string
		body         string
		serviceError bool
		want         want
	}{
		{
			name: "Positive counter",
			input: &models.Metrics{
				ID:    "counter-1",
				MType: "counter",
				Delta: ptrCount(1),
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Positive gauge",
			input: &models.Metrics{
				ID:    "gauge-1",
				MType: "gauge",
				Value: ptrGauge(12.3),
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:        "Invalid Content-Type",
			contentType: "text",
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "Invalid json",
			body: `{"id": "counter-2", "type": "counter", "value": 123,,}}`,
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "Service return error",
			input: &models.Metrics{
				ID:    "gauge-2",
				MType: "gauge",
				Value: ptrGauge(32.1),
			},
			want: want{
				code: http.StatusBadRequest,
			},
			serviceError: true,
		},
		{
			name: "Wrong type — value is string instead of float",
			body: `{"id": "temp", "type": "gauge", "value": "abc"}`,
			want: want{
				code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockMetricService(t)
			if tt.input != nil {
				var err error = nil
				if tt.serviceError {
					err = errors.New("err")
				}

				s.EXPECT().UpdateMetricJSON(context.Background(), *tt.input).
					Return(*tt.input, err)
			}

			w := httptest.NewRecorder()

			var reader io.Reader
			if tt.input != nil {
				buf := new(bytes.Buffer)
				err := json.NewEncoder(buf).Encode(tt.input)
				require.NoError(t, err)
				reader = buf
			} else {
				reader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/", reader)
			contentType := "application/json"
			if tt.contentType != "" {
				contentType = tt.contentType
			}
			req.Header.Add("Content-Type", contentType)

			h := PostJSON(s)
			h(w, req)

			resp := w.Result()
			assert.Equal(t, tt.want.code, resp.StatusCode)

			if tt.want.code != 200 || tt.input == nil {
				return
			}

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var got models.Metrics
			err = json.Unmarshal(respBody, &got)
			require.NoError(t, err)
			assert.Equal(t, *tt.input, got)
		})
	}
}

func ptrCount(val int64) *storage.Counter {
	res := storage.Counter(val)

	return &res
}

func ptrGauge(val float64) *storage.Gauge {
	res := storage.Gauge(val)

	return &res
}
