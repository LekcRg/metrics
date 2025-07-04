package value

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LekcRg/metrics/internal/merrors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	type want struct {
		contentType string
		value       string
		code        int
	}
	type metric struct {
		name  string
		mType string
		value string
	}
	tests := []struct {
		metric       metric
		name         string
		contentType  string
		want         want
		serviceError bool
	}{
		{
			name: "Valid gauge",
			metric: metric{
				name:  "one",
				mType: "gauge",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				value:       "12.34",
			},
		},
		{
			name: "Valid counter",
			metric: metric{
				name:  "one",
				mType: "counter",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				value:       "1234",
			},
		},
		{
			name:         "Service return error",
			serviceError: true,
			metric: metric{
				name:  "two",
				mType: "gauge",
			},
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockMetricGetter(t)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("name", tt.metric.name)
			rctx.URLParams.Add("type", tt.metric.mType)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
			r = r.WithContext(ctx)

			if (tt.want.code == http.StatusOK || tt.serviceError) &&
				tt.metric.name != "" && tt.metric.mType != "" {
				var err error = nil

				if tt.serviceError {
					err = merrors.ErrMocked
				}
				s.EXPECT().
					GetMetric(ctx, tt.metric.name, tt.metric.mType).
					Return(tt.want.value, err)
			}

			contentType := "text/plain"
			if tt.contentType != "" {
				contentType = tt.contentType
			}
			r.Header.Set("Content-Type", contentType)

			h := Get(s)
			h(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
