package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateSHA256(content []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(content)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}
