package dbping

import (
	"context"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type PingService struct {
	cfg config.ServerConfig
	db  storage.Storage
}

func NewPing(storage storage.Storage, cfg config.ServerConfig) *PingService {
	return &PingService{
		cfg: cfg,
		db:  storage,
	}
}

func (p PingService) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}
