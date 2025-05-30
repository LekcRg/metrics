package home

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

// Before
// BenchmarkGenerateHTML-8   1648   723813 ns/op   1927271 B/op   10022 allocs/op

// After
// BenchmarkGenerateHTML-8   3722   326983 ns/op   1098840 B/op   4006 allocs/op

func BenchmarkGenerateHTML(b *testing.B) {
	const lenList = 1000
	var (
		gaugeVal   = storage.Gauge(1234560.789)
		counterVal = storage.Counter(1234560789)
	)

	gaugeList := make(storage.GaugeCollection, lenList)
	counterList := make(storage.CounterCollection, lenList)
	for i := range lenList {
		gaugeList["gauge-"+strconv.Itoa(i)] = gaugeVal
		counterList["counter-"+strconv.Itoa(i)] = counterVal
	}

	b.ResetTimer()
	for range b.N {
		generateHTML(storage.Database{
			Gauge:   gaugeList,
			Counter: counterList,
		})
	}

	b.ReportAllocs()
}

func Test_generateHTML(t *testing.T) {
	type args struct {
		list storage.Database
	}
	tests := []struct {
		name         string
		db           storage.Database
		wantStatus   int
		wantContains []string
		wantErr      bool
	}{
		{
			name: "Normal metrics",
			db: storage.Database{
				Gauge:   map[string]storage.Gauge{"A": 1.1},
				Counter: map[string]storage.Counter{"B": 2},
			},
			wantStatus:   http.StatusOK,
			wantContains: []string{"A", "1.100", "B", "2"},
		},
		{
			name:         "Empty metrics",
			db:           storage.Database{},
			wantStatus:   http.StatusOK,
			wantContains: []string{"html"},
		},
		{
			name:       "Service error",
			db:         storage.Database{},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			svc := &fakeMetricService{
				db:      tt.db,
				wantErr: tt.wantErr,
			}
			handler := Get(svc)
			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			if !tt.wantErr {
				assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
			}

			body := w.Body.String()
			for _, substr := range tt.wantContains {
				assert.Contains(t, body, substr)
			}
		})
	}
}

type fakeMetricService struct {
	db      storage.Database
	wantErr bool
}

func (f *fakeMetricService) GetAllMetrics(ctx context.Context) (storage.Database, error) {
	var err error = nil
	if f.wantErr {
		err = errors.New("db err")
	}

	return f.db, err
}
