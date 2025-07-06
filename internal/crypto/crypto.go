package crypto

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var ErrPEMIsNil = errors.New("pem block is nil")

// GenerateHMAC возвращает строковое представление HMAC-SHA256 для заданных данных и ключа.
func GenerateHMAC(content []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(content)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}

func GetAndValidHMAC(ctx context.Context, key string, in proto.Message) error {
	if key == "" {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	headerHash := ""
	if ok {
		headerHash = md.Get("HashSHA256")[0]
	}

	if headerHash == "" {
		return status.Error(codes.PermissionDenied, "Empty HashSHA256")
	}

	inBytes, err := proto.Marshal(in)
	if err != nil {
		return status.Error(codes.Internal, "Internal server error")
	}

	hash := GenerateHMAC(inBytes, key)
	if hash != headerHash {
		return status.Error(codes.PermissionDenied, "Hash is not correct")
	}

	return nil
}
