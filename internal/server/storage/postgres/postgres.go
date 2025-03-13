package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

// TODO: Add contexts

func NewPostgres(config config.ServerConfig) (*Postgres, error) {
	conn, err := pgxpool.New(context.Background(), config.DatabaseDSN)
	if err != nil {
		logger.Log.Error("error while connecting to db")
		return nil, err
	}

	ctx := context.Background()

	_, err = conn.Exec(ctx, `create table if not exists gauge(
	name text not null unique PRIMARY KEY,
	value double precision not null,
	created_at timestamp with time zone not null default now()
	);`)
	if err != nil {
		logger.Log.Error(err.Error())
	}

	_, err = conn.Exec(ctx, `create table if not exists counter(
	name text not null unique PRIMARY KEY,
	value bigint not null,
	created_at timestamp with time zone not null default now()
	);`)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		db: conn,
	}, nil
}

func (p Postgres) UpdateCounter(name string, value storage.Counter) (storage.Counter, error) {
	req := `INSERT INTO counter (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = counter.value + $2
	RETURNING value;
	`
	row := p.db.QueryRow(context.Background(), req, name, value)

	var val sql.NullInt64
	err := row.Scan(&val)
	if err != nil {
		logger.Log.Error("error while scan setted counter value")
		return 0, err
	}

	if val.Valid {
		return storage.Counter(val.Int64), nil
	}

	return 0, fmt.Errorf("error while getting new value")
}

func (p Postgres) UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error) {
	req := `INSERT INTO gauge (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = EXCLUDED.value
	RETURNING value;
	`
	row := p.db.QueryRow(context.Background(), req, name, value)

	var val sql.NullFloat64
	err := row.Scan(&val)
	if err != nil {
		logger.Log.Error("error while scan setted counter value")
		return 0, err
	}

	if val.Valid {
		return storage.Gauge(val.Float64), nil
	}

	return 0, fmt.Errorf("error while getting new value")
}

func (p Postgres) UpdateMany(list storage.Database) error {
	reqCounter := `INSERT INTO counter (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = counter.value + $2
	RETURNING value;
	`
	reqGauge := `INSERT INTO gauge (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = EXCLUDED.value
	RETURNING value;
	`

	batch := &pgx.Batch{}

	for key, value := range list.Counter {
		batch.Queue(reqCounter, key, value)
	}

	for key, value := range list.Gauge {
		batch.Queue(reqGauge, key, value)
	}

	ctx := context.Background()
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	_, err = br.Exec()
	if err != nil {
		return err
	}

	err = br.Close()
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p Postgres) GetAllCounter() (storage.CounterCollection, error) {
	req := `SELECT name, value FROM counter`
	rows, err := p.db.Query(context.Background(), req)
	if err != nil {
		logger.Log.Error("error while sending request to db")
		return nil, err
	}

	list := make(storage.CounterCollection, 0)
	for rows.Next() {
		var name string
		var val sql.NullInt64
		err = rows.Scan(&name, &val)
		if err != nil {
			logger.Log.Error(err.Error())
			return nil, err
		}

		if !val.Valid {
			return nil, fmt.Errorf("error while validate value")
		}

		list[name] = storage.Counter(val.Int64)
	}
	return list, nil
}

func (p Postgres) GetAllGouge() (storage.GaugeCollection, error) {
	req := `SELECT name, value FROM gauge`
	rows, err := p.db.Query(context.Background(), req)
	if err != nil {
		logger.Log.Error("error while sending request to db")
		return nil, err
	}

	list := make(storage.GaugeCollection, 0)
	for rows.Next() {
		var name string
		var val sql.NullFloat64
		err = rows.Scan(&name, &val)
		if err != nil {
			logger.Log.Error(err.Error())
			return nil, err
		}

		if !val.Valid {
			return nil, fmt.Errorf("error while validate value")
		}

		list[name] = storage.Gauge(val.Float64)
	}
	return list, nil
}

func (p Postgres) GetGaugeByName(name string) (storage.Gauge, error) {
	req := `SELECT value FROM gauge WHERE name=$1 LIMIT 1`
	row := p.db.QueryRow(context.Background(), req, name)

	var val sql.NullFloat64
	row.Scan(&val)

	if !val.Valid {
		logger.Log.Info("found null element")
		return 0, fmt.Errorf("not found")
	}

	return storage.Gauge(val.Float64), nil
}

func (p Postgres) GetCounterByName(name string) (storage.Counter, error) {
	req := `SELECT value FROM counter WHERE name=$1 LIMIT 1`
	row := p.db.QueryRow(context.Background(), req, name)

	var val sql.NullInt64
	row.Scan(&val)

	if !val.Valid {
		logger.Log.Info("found null element")
		return 0, fmt.Errorf("not found")
	}

	return storage.Counter(val.Int64), nil
}

func (p Postgres) GetAll() (storage.Database, error) {
	gaugeList, err := p.GetAllGouge()
	if err != nil {
		logger.Log.Error(err.Error())
		return storage.Database{}, err
	}

	counterList, err := p.GetAllCounter()
	if err != nil {
		logger.Log.Error(err.Error())
		return storage.Database{}, err
	}

	return storage.Database{
		Gauge:   gaugeList,
		Counter: counterList,
	}, nil
}

func (p Postgres) Ping() error {
	if p.db != nil {
		return p.db.Ping(context.Background())
	} else {
		return fmt.Errorf("db is not connected")
	}
}

func (p Postgres) Close() {
	p.db.Close()
}
