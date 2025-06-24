package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrPEMIsNil = errors.New("pem block is nil")

// GenerateHMAC возвращает строковое представление HMAC-SHA256 для заданных данных и ключа.
func GenerateHMAC(content []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(content)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}
