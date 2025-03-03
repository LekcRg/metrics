package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LekcRg/metrics/internal/server/services/metric"
	"github.com/LekcRg/metrics/internal/server/services/store"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/LekcRg/metrics/internal/server/storage/memstorage"
	"github.com/LekcRg/metrics/internal/testdata"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueRoutes(t *testing.T) {
	valueStorage, _ := memstorage.New()
	config := testdata.TestServerConfig
	store := store.NewStore(valueStorage, config)
	updateService := metric.NewMetricsService(valueStorage, config, store)
	r := chi.NewRouter()
	ValueRoutes(r, *updateService)
	ts := httptest.NewServer(r)
	defer ts.Close()
	valueStorage.UpdateGauge("one", storage.Gauge(123.45))
	valueStorage.UpdateCounter("two", storage.Counter(12345))
	valueStorage.UpdateGauge("five", storage.Gauge(-123.45))
	valueStorage.UpdateCounter("six", storage.Counter(-12345))

	type want struct {
		code        int
		contentType string
		response    string
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "#1[GET] Positive gauge number",
			url:  "/value/gauge/one",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "123.45",
			},
		},
		{
			name: "#2[GET] Positive gauge number",
			url:  "/value/counter/two",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "12345",
			},
		},
		{
			name: "#3[GET] Negative gauge number",
			url:  "/value/gauge/five",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "-123.45",
			},
		},
		{
			name: "#4[GET] Positive gauge number",
			url:  "/value/counter/six",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				response:    "-12345",
			},
		},
		{
			name: "#5[GET] Counter not found",
			url:  "/value/counter/three",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				response:    "",
			},
		},
		{
			name: "#6[GET] Gauge not found",
			url:  "/value/counter/four",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				response:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, ts.URL+tt.url, nil)
			require.NoError(t, err)
			resp, err := ts.Client().Do(req)

			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			if resp.StatusCode == 200 {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(body))
			}
		})
	}
}
