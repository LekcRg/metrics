package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LekcRg/metrics/internal/server/storage/memStorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {

	routerStorage, _ := memStorage.New()
	r := NewRouter(routerStorage)
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
			name: "#1 Get request /",
			url:  "/",
			want: want{
				code:        http.StatusOK,
				contentType: "text/html",
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
		})
	}

}
