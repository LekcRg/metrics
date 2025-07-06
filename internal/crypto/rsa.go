package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/LekcRg/metrics/internal/logger"
	"go.uber.org/zap"
)

type RsaKey interface {
	Size() int

	*rsa.PrivateKey | *rsa.PublicKey
}

func EncryptRSA(data []byte, key *rsa.PublicKey) ([]byte, error) {
	res := make([]byte, 0, len(data))

	maxChunkSize := key.Size() - 11 // PKCS1v15 padding

	for i := 0; i < len(data); i += maxChunkSize {
		end := min(i+maxChunkSize, len(data))

		chunk := data[i:end]
		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, key, chunk)
		if err != nil {
			return nil, err
		}
		res = append(res, encrypted...)
	}

	return res, nil
}

func DecryptRSA(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	res := make([]byte, 0, len(data))

	blockSize := key.Size()

	for i := 0; i < len(data); i += blockSize {
		if i+blockSize > len(data) {
			return nil, fmt.Errorf("invalid encrypted data length")
		}

		chunk := data[i : i+blockSize]
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, key, chunk)
		if err != nil {
			return nil, err
		}
		res = append(res, decrypted...)
	}

	return res, nil
}

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

			newBody, err := DecryptRSA(body, priv)
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
