package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		body   []byte
		code   int
	}{
		{
			name:   "Get status 200",
			code:   http.StatusOK,
			body:   []byte("test1"),
			url:    "/test/1",
			method: http.MethodGet,
		},
		{
			name:   "POST status 400",
			code:   http.StatusBadRequest,
			body:   []byte("2test"),
			url:    "/test/second",
			method: http.MethodPost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.InfoLevel)
			Log = zap.New(core)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
				w.Write(tt.body)
			})

			wrapped := RequestLogger(testHandler)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			require.Equal(t, tt.code, rec.Code)
			require.Equal(t, tt.body, rec.Body.Bytes())

			entries := logs.All()
			require.Len(t, entries, 1)

			entry := entries[0]
			assert.Equal(t, "got incoming HTTP request", entry.Message)

			fields := map[string]zapcore.Field{}
			for _, f := range entry.Context {
				fields[f.Key] = f
			}

			assert.Equal(t, tt.method, fields["method"].String)
			assert.Equal(t, tt.url, fields["path"].String)
			assert.Equal(t, int64(tt.code), fields["status"].Integer)
			assert.Equal(t, int64(len(tt.body)), fields["size"].Integer) // len("OK")
			assert.True(t, fields["time"].Type == zapcore.DurationType)
		})
	}
}

func TestInitialize(t *testing.T) {
	type args struct {
		level string
		isDev bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid",
			args: args{
				level: "debug",
				isDev: false,
			},
		},
		{
			name:    "Init with invalid level",
			wantErr: true,
			args: args{
				level: "invalid",
				isDev: false,
			},
		},
		{
			name:    "Dev mode",
			wantErr: false,
			args: args{
				level: "debug",
				isDev: true,
			},
		},
		{
			name:    "Info level",
			wantErr: false,
			args: args{
				level: "info",
				isDev: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.args.level, tt.args.isDev)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NotNil(t, Log)
		})
	}
}
