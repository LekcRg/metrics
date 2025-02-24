package cgzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
)

type (
	gzipWriter struct {
		http.ResponseWriter
		headerData *headerData
	}

	headerData struct {
		statusCode int
	}
)

var toGzip = []string{
	"application/javascript",
	"application/json",
	"text/css",
	"text/html",
	"text/plain",
	"text/xml",
}

func (w gzipWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")

	if !slices.Contains(toGzip, contentType) || w.headerData.statusCode > 299 {
		w.WriteHeader(w.headerData.statusCode)
		return w.ResponseWriter.Write(b)
	}

	gz, err := gzip.NewWriterLevel(w.ResponseWriter, gzip.BestSpeed)
	if err != nil {
		io.WriteString(w, err.Error())
	}
	defer gz.Close()

	w.Header().Add("Content-Encoding", "gzip")
	if w.headerData.statusCode > 0 {
		w.ResponseWriter.WriteHeader(w.headerData.statusCode)
	}
	return gz.Write(b)
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

		next.ServeHTTP(gzipWriter{ResponseWriter: w, headerData: headerData}, r)
	})
}
