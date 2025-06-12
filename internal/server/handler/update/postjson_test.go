package update

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	counter1 = models.Metrics{
		ID:    "counter-1",
		MType: "counter",
		Delta: ptrCount(1),
	}
	counter2 = models.Metrics{
		ID:    "counter-2",
		MType: "counter",
		Delta: ptrCount(2),
	}
	gauge1 = models.Metrics{
		ID:    "gauge-1",
		MType: "gauge",
		Value: ptrGauge(12.3),
	}
	gauge2 = models.Metrics{
		ID:    "gauge-2",
		MType: "gauge",
		Value: ptrGauge(32.1),
	}
)

func TestPostJSON(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		input        *models.Metrics
		name         string
		contentType  string
		body         string
		want         want
		serviceError bool
	}{
		{
			name:  "Positive counter",
			input: &counter1,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:  "Positive gauge",
			input: &gauge1,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:        "Invalid Content-Type",
			contentType: "text",
			want: want{
				code: http.StatusBadRequest,
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
			name:  "Service return error",
			input: &gauge2,
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
			s := NewMockMetricUpdater(t)
			if tt.input != nil {
				var err error = nil
				if tt.serviceError {
					err = merrors.ErrMocked
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
			defer resp.Body.Close()
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

func TestPostMany(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name         string
		contentType  string
		body         string
		SHA256       string
		key          string
		input        []models.Metrics
		want         want
		serviceError bool
	}{
		{
			name: "Change one counter",
			input: []models.Metrics{
				counter1,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Change two counters",
			input: []models.Metrics{
				counter1,
				counter2,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Change one gauge",
			input: []models.Metrics{
				gauge1,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Change two gauges",
			input: []models.Metrics{
				gauge1,
				gauge2,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Change gauge and counter",
			input: []models.Metrics{
				gauge1,
				counter1,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Change 4 other elements",
			input: []models.Metrics{
				gauge1,
				counter1,
				gauge2,
				counter2,
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Invalid json",
			body: `[
				{"id": "counter-2", "type": "counter", "value": 123},
				{"id": "counter-1", "type": "counter", "value": 123}},
			]`,
			want: want{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "Error from service",
			input: []models.Metrics{
				counter1,
			},
			want: want{
				code: http.StatusInternalServerError,
			},
			serviceError: true,
		},
		{
			name: "Valid with SHA256",
			input: []models.Metrics{
				counter2,
				counter1,
			},
			want: want{
				code: http.StatusOK,
			},
			key: "test-key",
		},
		{
			name:   "Invalid SHA256 string",
			SHA256: `invalid`,
			want: want{
				code: http.StatusBadRequest,
			},
			key: "test-key",
		},
		{
			name: "With key without SHA256 header",
			want: want{
				code: http.StatusBadRequest,
			},
			key: "test",
		},
		{
			name:        "Invalid Content-Type",
			contentType: "text",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockMetricUpdater(t)
			if len(tt.input) > 0 {
				var err error = nil

				if tt.serviceError {
					err = merrors.ErrMocked
				}
				s.EXPECT().UpdateMany(context.Background(), tt.input).
					Return(err)
			}

			w := httptest.NewRecorder()

			var reader io.Reader
			buf := new(bytes.Buffer)
			if tt.input != nil {
				err := json.NewEncoder(buf).Encode(tt.input)
				require.NoError(t, err)
				reader = buf
			} else {
				reader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/", reader)

			if tt.key != "" {
				sha := tt.SHA256
				if sha == "" && tt.input != nil {
					sha = crypto.GenerateHMAC(buf.Bytes(), tt.key)
				}

				req.Header.Add("HashSHA256", sha)
			}

			contentType := "application/json"
			if tt.contentType != "" {
				contentType = tt.contentType
			}
			req.Header.Add("Content-Type", contentType)

			h := PostMany(s, tt.key)
			h(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode)

			if tt.want.code != 200 || tt.input == nil {
				return
			}
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
