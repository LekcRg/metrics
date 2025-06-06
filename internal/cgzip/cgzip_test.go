// Package cgzip предоставляет middleware и утилиты для работы с GZIP-сжатием в HTTP-запросах и ответах.
package cgzip

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const content = "content"

func getHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(statusCode)
		w.Write([]byte(content))
	})
}

func TestGzipHandle(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		wantCode      int
		isGzip        bool
		withoutAccept bool
	}{
		{
			name:       "gzipped 200",
			statusCode: http.StatusOK,
			wantCode:   http.StatusOK,
			isGzip:     true,
		},
		{
			name:     "gzipped without status",
			wantCode: http.StatusOK,
			isGzip:   true,
		},
		{
			name:       "not gzipped 404",
			statusCode: http.StatusNotFound,
			wantCode:   http.StatusNotFound,
		},
		{
			name:          "not gzipped 200 without Accept-Encoding",
			statusCode:    http.StatusOK,
			wantCode:      http.StatusOK,
			withoutAccept: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			if !tt.withoutAccept {
				r.Header.Add("Accept-Encoding", "gzip")
			}
			h := GzipHandle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(content))
			}))
			h.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantCode, res.StatusCode)

			body := ungzip(t, res, tt.isGzip)
			assert.Equal(t, content, string(body))
		})
	}
}

func TestGzipBody(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		wantCode        int
		isGzip          bool
		withoutEncoding bool
	}{
		{
			name:   "gzipped",
			isGzip: true,
		},
		{
			name:            "without Content-Encoding without gzipped body",
			isGzip:          false,
			withoutEncoding: true,
		},
		{
			name:   "with Content-Encoding without gzipped body",
			isGzip: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := GzipBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, content, string(body))

				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(content))
			}))

			var buf bytes.Buffer
			if tt.isGzip {
				getGzippedContent(t, &buf)
			} else {
				buf.WriteString(content)
			}

			r := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(&buf))
			if !tt.withoutEncoding {
				r.Header.Add("Content-Encoding", "gzip")
			}

			h.ServeHTTP(w, r)
		})
	}
}

func TestGetGzippedReq(t *testing.T) {
	tests := []struct {
		name    string
		content string
		url     string
		wantErr bool
	}{
		{
			name:    "gzipped content",
			content: content,
		},
		{
			name: "gzipped empty line",
		},
		{
			name:    "invalid url",
			content: content,
			url:     ":",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gzippedBuf bytes.Buffer
			getGzipped(t, &gzippedBuf, tt.content)
			gzippedContent, err := io.ReadAll(&gzippedBuf)
			require.NoError(t, err)

			r, err := GetGzippedReq(context.Background(), tt.url, []byte(tt.content))
			if tt.wantErr {
				require.Error(t, err)

				return
			} else {
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, gzippedContent, body)

				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("content"))
			})

			h.ServeHTTP(w, r)
		})
	}
}
func ungzip(t *testing.T, res *http.Response, isGzip bool) []byte {
	var body []byte

	if isGzip {
		gz, err := gzip.NewReader(res.Body)
		require.NoError(t, err)
		defer gz.Close()

		body, err = io.ReadAll(gz)
		require.NoError(t, err)
	} else {
		var err error
		body, err = io.ReadAll(res.Body)
		require.NoError(t, err)
	}

	return body
}

func getGzipped(t *testing.T, buf *bytes.Buffer, str string) {
	gz, err := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	require.NoError(t, err)

	_, err = gz.Write([]byte(str))
	require.NoError(t, err)

	require.NoError(t, gz.Close())
}

func getGzippedContent(t *testing.T, buf *bytes.Buffer) {
	getGzipped(t, buf, content)
}
