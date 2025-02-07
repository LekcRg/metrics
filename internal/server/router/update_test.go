package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LekcRg/metrics/internal/server/storage/memstorage"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRoutes(t *testing.T) {
	updateStorage, _ := memstorage.New()
	r := chi.NewRouter()
	UpdateRoutes(r, updateStorage)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "#1[POST] Positive gauge number with point",
			url:  "/update/gauge/one/12.34",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#2[POST] Positive gauge number without point",
			url:  "/update/gauge/two/5678",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#3[POST] Positive counter",
			url:  "/update/counter/three/3",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#4[POST] Positive counter second request",
			url:  "/update/counter/three/3",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#5[POST] Positive counter third request with other num",
			url:  "/update/counter/three/1",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#6[POST] Positive gauge number with the same name as the previous one",
			url:  "/update/gauge/one/12",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#7[POST] Negative request without value",
			url:  "/update/gauge/one",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#8[POST] Negative request without value and name",
			url:  "/update/gauge",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#9[POST] Negative request without params",
			url:  "/update",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#10[POST] Negative request with wrong gauge value",
			url:  "/update/gauge/nine/one_point_two",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#11[POST] Negative request with wrong counter value",
			url:  "/update/counter/ten/twelve",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "#12[POST] Negative request with wrong type",
			url:  "/update/integer/eleven/2025",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL+tt.url, nil)
			require.NoError(t, err)

			req.Header.Add("Content-Type", "text/plain")
			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			resp.Body.Close()
		})
	}
}
