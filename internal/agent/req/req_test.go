package req

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPRequest(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		resStatus int
		retry     int
		wantErr   bool
	}{
		{
			name:      "Success response",
			resStatus: http.StatusOK,
		},
		{
			name:      "Error response",
			resStatus: http.StatusInternalServerError,
			wantErr:   true,
		},
		{
			name:      "Test SHA256 encryption",
			key:       "test",
			resStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			val := storage.Gauge(1.234)
			metrics := []models.Metrics{
				{
					ID:    "test",
					MType: "gauge",
					Value: &val,
				},
			}
			body, err := json.Marshal(metrics)
			require.NoError(t, err)
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.key != "" {
					sha := crypto.GenerateHMAC(body, tt.key)

					hash := r.Header.Get("HashSHA256")
					assert.Equal(t, sha, hash)
				}

				w.WriteHeader(tt.resStatus)
				w.Write([]byte("test"))
			}))

			err = HTTPRequest(RequestArgs{
				Ctx:     ctx,
				URL:     svr.URL,
				Metrics: metrics,
				Config: config.AgentConfig{
					CommonConfig: config.CommonConfig{
						Key: tt.key,
					},
				},
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
