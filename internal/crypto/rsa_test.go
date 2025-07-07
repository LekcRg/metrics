package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pathToPriv       = "../testdata/keys/priv.pem"
	pathToPub        = "../testdata/keys/pub.pem"
	pathToInvalidPEM = "../testdata/invalid.pem"
)

func TestRsaMiddleware(t *testing.T) {
	privPath, err := filepath.Abs(pathToPriv)
	t.Log(privPath)
	require.NoError(t, err)
	pemPriv, err := ParsePEMFile(privPath)
	require.NoError(t, err)
	priv, err := x509.ParsePKCS1PrivateKey(pemPriv)
	require.NoError(t, err)

	pubPath, err := filepath.Abs(pathToPub)
	require.NoError(t, err)
	pemPub, err := ParsePEMFile(pubPath)
	require.NoError(t, err)
	pub, err := x509.ParsePKCS1PublicKey(pemPub)
	require.NoError(t, err)

	tests := []struct {
		key          *rsa.PrivateKey
		name         string
		msg          string
		method       string
		encryptedMsg []byte
		code         int
		isEncrypted  bool
		notCheckBody bool
		withoutKey   bool
	}{
		{
			name:        "Encrypted POST",
			msg:         "Test encryption",
			method:      http.MethodPost,
			isEncrypted: true,
		},
		{
			name:         "Invalid encrypted POST",
			encryptedMsg: []byte("invalid encryption"),
			method:       http.MethodPost,
			code:         http.StatusInternalServerError,
		},
		{
			name:         "Not encrypted POST",
			msg:          "Test encryption",
			method:       http.MethodPost,
			isEncrypted:  false,
			notCheckBody: true,
			code:         http.StatusInternalServerError,
		},
		{
			name:        "Not encrypted GET",
			msg:         "Test encryption",
			method:      http.MethodGet,
			isEncrypted: false,
		},
		{
			name:        "Without key not encrypted POST",
			msg:         "Test encryption",
			method:      http.MethodPost,
			isEncrypted: false,
			withoutKey:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			msg := []byte(tt.msg)
			if tt.encryptedMsg != nil || len(tt.encryptedMsg) > 0 {
				msg = tt.encryptedMsg
			} else if tt.isEncrypted {
				msg, err = rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(tt.msg))
				require.NoError(t, err)
			}

			code := tt.code
			if code == 0 {
				code = http.StatusOK
			}

			w := httptest.NewRecorder()
			m := RsaMiddleware(priv)
			if tt.withoutKey {
				m = RsaMiddleware(nil)
			}
			h := m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				defer r.Body.Close()

				if !tt.notCheckBody {
					assert.Equal(t, tt.msg, string(body))
				}

				w.Header().Set("Content-Type", "text")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			var buf bytes.Buffer
			buf.Write(msg)

			r := httptest.NewRequest(tt.method, "/", io.NopCloser(&buf))

			h.ServeHTTP(w, r)
		})
	}
}

func TestParsePEMFile(t *testing.T) {
	validPath, err := filepath.Abs(pathToPriv)
	require.NoError(t, err)
	invalidPath, err := filepath.Abs("invalid-path.pem")
	require.NoError(t, err)
	invalidFilePath, err := filepath.Abs(pathToInvalidPEM)
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name: "valid path",
			path: validPath,
		},
		{
			name:    "invalid path",
			path:    invalidPath,
			wantErr: true,
		},
		{
			name:    "invalid PEM",
			path:    invalidFilePath,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePEMFile(tt.path)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEncryptDecryptRSA(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	tests := [][]byte{
		[]byte("Hello, World!"),
		[]byte(""), // пустые данные
		[]byte(strings.Repeat("long message", 50)),
	}

	for _, original := range tests {
		encrypted, err := EncryptRSA(original, publicKey)
		require.NoError(t, err)

		decrypted, err := DecryptRSA(encrypted, privateKey)
		require.NoError(t, err)

		assert.Equal(t, original, decrypted)
	}
}
