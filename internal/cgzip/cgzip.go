package cgzip

import (
	"compress/gzip"
	"net/http"
	"slices"
	"strings"
	"sync"
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
	if !slices.Contains(toGzip, contentType) ||
		w.headerData.statusCode > 299 ||
		len(b) < 1400 {
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
