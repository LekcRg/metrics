package cgzip

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/LekcRg/metrics/internal/logger"
)

var gzwrPool = &sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

type (
	gzipWriter struct {
		http.ResponseWriter
		Writer     *gzip.Writer
		headerData *headerData
	}

	headerData struct {
		statusCode int
	}
)

var toGzip = []string{
	"application/json",
	"text/html",
	// "application/javascript",
	// "text/css",
	// "text/plain",
	// "text/xml",
}

func (w gzipWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if w.headerData.statusCode == 0 {
		w.headerData.statusCode = http.StatusOK
	}

	// TODO: Add check for the content-type with utf-8, etc.
	// Can check len(b) < 1400
	if !slices.Contains(toGzip, contentType) ||
		w.headerData.statusCode > 299 {
		w.ResponseWriter.WriteHeader(w.headerData.statusCode)
		return w.ResponseWriter.Write(b)
	}

	gzw := gzwrPool.Get().(*gzip.Writer)
	gzw.Reset(w.ResponseWriter)
	w.Writer = gzw
	w.Header().Add("Content-Encoding", "gzip")
	w.Header().Del("Content-Length")
	if w.headerData.statusCode > 0 {
		w.ResponseWriter.WriteHeader(w.headerData.statusCode)
	}

	defer func() {
		w.Writer.Close()
		gzwrPool.Put(w.Writer)
	}()
	return w.Writer.Write(b)
}

func (w gzipWriter) WriteHeader(statusCode int) {
	w.headerData.statusCode = statusCode
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		headerData := &headerData{statusCode: 0}
		gzwr := gzipWriter{
			ResponseWriter: w,
			headerData:     headerData,
		}

		next.ServeHTTP(gzwr, r)
	})
}

func GzipBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Log.Error("Error while create gzip reader")
			}

			r.Body = gz
			defer gz.Close()
		}

		next.ServeHTTP(w, r)
	})
}

// add context in the future, idk why)
func GetGzippedReq(url string, body []byte) (*http.Request, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		logger.Log.Error("Error creating gzip writer")
		return nil, err
	}
	_, err = gz.Write(body)
	if err != nil {
		logger.Log.Error("Error writing to gzip writer")
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		logger.Log.Error("Error closing gzip writer")
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		logger.Log.Error("Error creating http request")
		return nil, err
	}
	req.Header.Set("Content-Encoding", "gzip")
	return req, nil
}
