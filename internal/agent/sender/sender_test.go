package sender

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSender_postRequest(t *testing.T) {
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
		{
			name:      "Retry 1 time",
			retry:     1,
			resStatus: http.StatusOK,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			body := []byte("test")
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.key != "" {
					sha := crypto.GenerateHMAC(body, tt.key)

					hash := r.Header.Get("HashSHA256")
					assert.Equal(t, sha, hash)
				}

				if tt.retry > 0 {
					cancel()
					tt.retry--
				}
				w.WriteHeader(tt.resStatus)
				w.Write([]byte("test"))
			}))
			s := &Sender{
				url:     svr.URL,
				monitor: &monitoring.MonitoringStats{},
				config: config.AgentConfig{
					CommonConfig: config.CommonConfig{
						Key: tt.key,
					},
				},
			}

			err := s.postRequest(ctx, body)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
