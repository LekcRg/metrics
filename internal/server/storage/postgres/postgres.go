package postgres

import (
	"context"
	"fmt"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/jackc/pgx/v5"
)

type Postgres struct {
	db *pgx.Conn
}

func NewPostgres(config config.ServerConfig) *Postgres {
	conn, err := pgx.Connect(context.Background(), config.DatabaseDSN)
	if err != nil {
		logger.Log.Error("error while connecting to db")
	}

	return &Postgres{
		db: conn,
	}
}

func (p *Postgres) Ping() error {
	if p.db != nil {
		return p.db.Ping(context.Background())
	} else {
		return fmt.Errorf("db is not connected")
	}
}
