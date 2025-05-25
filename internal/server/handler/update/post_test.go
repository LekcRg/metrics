package update

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	type metric struct {
		name  string
		mType string
		value string
	}
	tests := []struct {
		name         string
		contentType  string
		serviceError bool
		metric       metric
		want         want
	}{
		{
			name: "Valid gauge",
			metric: metric{
				name:  "one",
				mType: "gauge",
				value: "12.34",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Valid counter",
			metric: metric{
				name:  "one",
				mType: "counter",
				value: "12",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "Incorrect Content-Type",
			contentType: "application/json",
			metric: metric{
				name:  "two",
				mType: "gauge",
				value: "2",
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:         "Service return error",
			serviceError: true,
			metric: metric{
				name:  "two",
				mType: "gauge",
				value: "2",
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMockMetricUpdater(t)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("name", tt.metric.name)
			rctx.URLParams.Add("type", tt.metric.mType)
			rctx.URLParams.Add("value", tt.metric.value)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)

			ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
			r = r.WithContext(ctx)

			if (tt.want.code == http.StatusOK || tt.serviceError) &&
				tt.metric.name != "" && tt.metric.mType != "" && tt.metric.value != "" {
				var err error = nil

				if tt.serviceError {
					err = errors.New("err")
				}
				s.EXPECT().
					UpdateMetric(ctx, tt.metric.name, tt.metric.mType, tt.metric.value).
					Return(err)
			}

			contentType := "text/plain"
			if tt.contentType != "" {
				contentType = tt.contentType
			}
			r.Header.Set("Content-Type", contentType)

			h := Post(s)
			h(w, r)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
