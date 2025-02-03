package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/LekcRg/metrics/internal/server/storage/memStorage"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRoutes(t *testing.T) {
	updateStorage, _ := memStorage.New()
	r := chi.NewRouter()
	UpdateRoutes(r, updateStorage)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type wantDb struct {
		vType string
		name  string
		value any
		check bool
	}
	type want struct {
		code        int
		contentType string
		db          wantDb
	}
	tests := []struct {
		name        string
		url         string
		contentType string
		want        want
	}{
		// TODO: Add test cases.
		{
			name: "#1 Positive gauge number with point",
			url:  "/update/gauge/one/12.34",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					vType: "gauge",
					name:  "one",
					value: storage.Gauge(12.34),
					check: true,
				},
			},
		},
		{
			name: "#2 Positive gauge number without point",
			url:  "/update/gauge/two/5678",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					vType: "gauge",
					name:  "two",
					value: storage.Gauge(5678),
					check: true,
				},
			},
		},
		{
			name: "#3 Positive counter",
			url:  "/update/counter/three/3",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					vType: "counter",
					name:  "three",
					value: storage.Counter(3),
					check: true,
				},
			},
		},
		{
			name: "#4 Positive counter second request",
			url:  "/update/counter/three/3",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					vType: "counter",
					name:  "three",
					value: storage.Counter(6),
					check: true,
				},
			},
		},
		{
			name: "#5 Positive counter third request with other num",
			url:  "/update/counter/three/1",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					vType: "counter",
					name:  "three",
					value: storage.Counter(7),
					check: true,
				},
			},
		},
		{
			name: "#6 Positive gauge number with the same name as the previous one",
			url:  "/update/gauge/one/12",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					vType: "gauge",
					name:  "one",
					value: storage.Gauge(12),
					check: true,
				},
			},
		},
		{
			name: "#7 Negative request without value",
			url:  "/update/gauge/one",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
		{
			name: "#8 Negative request without value and name",
			url:  "/update/gauge",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
		{
			name: "#9 Negative request without params",
			url:  "/update",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
		{
			name: "#10 Negative request with wrong gauge value",
			url:  "/update/gauge/nine/one_point_two",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
		{
			name: "#11 Negative request with wrong counter value",
			url:  "/update/counter/ten/twelve",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
		{
			name: "#12 Negative request with wrong type",
			url:  "/update/integer/eleven/2025",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
		{
			name:        "#13 Negative request with wrong content-type",
			url:         "/update/integer/eleven/2025",
			contentType: "multipart/form-data",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				db: wantDb{
					check: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL+tt.url, nil)
			require.NoError(t, err)

			if tt.contentType != "" {
				req.Header.Add("Content-Type", tt.contentType)
			} else {
				req.Header.Add("Content-Type", "text/plain")
			}

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))

			if tt.want.db.check {
				if tt.want.db.vType == "gauge" {
					value, err := updateStorage.GetGaugeByName(tt.want.db.name)

					if err != nil {
						t.Errorf("Not saved to db")
					} else {
						assert.Equal(t, value, tt.want.db.value)
					}
				} else if tt.want.db.vType == "counter" {
					value, err := updateStorage.GetCounterByName(tt.want.db.name)

					if err != nil {
						t.Errorf("Not saved to db")
					} else {
						assert.Equal(t, value, tt.want.db.value)
					}
				}
			}
		})
	}
}
