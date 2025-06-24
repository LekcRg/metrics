package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/LekcRg/metrics/internal/logger"
	"go.uber.org/zap"
)

func ParsePEMFile(path string) ([]byte, error) {
	publicFile, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = publicFile.Close()
		if err != nil {
			logger.Log.Error("Error defer close pem key file", zap.Error(err))
		}
	}()

	data, err := io.ReadAll(publicFile)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(data))
	if keyBlock == nil {
		return nil, ErrPEMIsNil
	}

	return keyBlock.Bytes, nil
}

func RsaMiddleware(priv *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if priv == nil || r.Method != http.MethodPost {
				next.ServeHTTP(w, r)

				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Error while read body with rsa")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			defer r.Body.Close()

			newBody, err := rsa.DecryptPKCS1v15(rand.Reader, priv, body)
			if err != nil {
				logger.Log.Error("Error while decrypt body")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}

			r.Body = io.NopCloser(bytes.NewReader(newBody))
			r.ContentLength = int64(len(newBody))
			next.ServeHTTP(w, r)
		})
	}
}
