package dbping

import "github.com/LekcRg/metrics/internal/config"

type db interface {
	Ping() error
}

type PingService struct {
	cfg config.ServerConfig
	db  db
}

func NewPing(storage db, cfg config.ServerConfig) *PingService {
	return &PingService{
		cfg: cfg,
		db:  storage,
	}
}

func (p PingService) Ping() error {
	return p.db.Ping()
}
