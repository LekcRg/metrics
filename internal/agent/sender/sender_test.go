package sender

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LekcRg/metrics/internal/agent/monitoring"
	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSender(t *testing.T) {
	received := []models.Metrics{}
	var mu sync.Mutex
	reqCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			logger.Log.Error("Error while create gzip reader")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		var metrics []models.Metrics
		err = json.NewDecoder(gz).Decode(&metrics)
		require.NoError(t, err)

		mu.Lock()
		received = append(received, metrics...)
		reqCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	mon := monitoring.New(1)

	sender := New(config.AgentConfig{
		RateLimit:      5,
		ReportInterval: 1,
		Addr:           strings.TrimPrefix(ts.URL, "http://"),
	}, mon, nil)

	sender.Start(ctx, &wg)
	mon.Start(ctx, &wg)

	time.Sleep(5 * time.Second)

	sender.Shutdown()
	mon.Shutdown()

	assert.Greater(t, reqCount, 5, "request lower than 6")

	needMetric := []string{"PollCount", "RandomValue", "Alloc", "TotalMemory"}
	countFound := 0
	for _, m := range received {
		if slices.Contains(needMetric, m.ID) {
			countFound++
		}
	}
	assert.Greater(t, countFound, len(needMetric)-1)
}
