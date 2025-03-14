package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LekcRg/metrics/internal/common"
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
		return nil, err
	}

	err = common.Retry(func() error {
		err = conn.Ping(context.Background())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	err = common.Retry(func() error {
		_, err := conn.Exec(ctx, `create table if not exists gauge(
		name text not null unique PRIMARY KEY,
		value double precision not null,
		created_at timestamp with time zone not null default now()
		);`)

		return err
	})

	if err != nil {
		return nil, err
	}

	err = common.Retry(func() error {
		_, err = conn.Exec(ctx, `create table if not exists counter(
		name text not null unique PRIMARY KEY,
		value bigint not null,
		created_at timestamp with time zone not null default now()
		);`)

		return err
	})

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
	var result storage.Counter

	err := common.Retry(func() error {
		row := p.db.QueryRow(context.Background(), req, name, value)

		var val sql.NullInt64
		err := row.Scan(&val)
		if err != nil {
			logger.Log.Error("error while scan setted counter value")
			return err
		}

		if val.Valid {
			result = storage.Counter(val.Int64)
			return nil
		}

		return fmt.Errorf("error while getting new value")
	})

	if err != nil {
		return 0, err
	}

	return result, nil
}

func (p Postgres) UpdateGauge(name string, value storage.Gauge) (storage.Gauge, error) {
	req := `INSERT INTO gauge (name, value)
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE
	SET value = EXCLUDED.value
	RETURNING value;
	`
	var result storage.Gauge

	err := common.Retry(func() error {
		row := p.db.QueryRow(context.Background(), req, name, value)

		var val sql.NullFloat64
		err := row.Scan(&val)
		if err != nil {
			logger.Log.Error("error while scan setted gauge value")
			return err
		}

		if val.Valid {
			result = storage.Gauge(val.Float64)
			return nil
		}

		return fmt.Errorf("error while getting new value")
	})

	if err != nil {
		return 0, err
	}

	return result, nil
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
	return common.Retry(func() error {
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
		return err
	})
}

func (p Postgres) GetAllCounter() (storage.CounterCollection, error) {
	req := `SELECT name, value FROM counter`
	var list storage.CounterCollection
	err := common.Retry(func() error {
		rows, err := p.db.Query(context.Background(), req)
		if err != nil {
			logger.Log.Error("error while sending request to db")
			return err
		}
		defer rows.Close()

		list = make(storage.CounterCollection, 0)
		for rows.Next() {
			var name string
			var val sql.NullInt64
			err = rows.Scan(&name, &val)
			if err != nil {
				logger.Log.Error(err.Error())
				return err
			}

			if !val.Valid {
				return fmt.Errorf("error while validate value")
			}

			list[name] = storage.Counter(val.Int64)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (p Postgres) GetAllGauge() (storage.GaugeCollection, error) {
	req := `SELECT name, value FROM gauge`

	var list storage.GaugeCollection
	err := common.Retry(func() error {
		rows, err := p.db.Query(context.Background(), req)
		if err != nil {
			logger.Log.Error("error while sending request to db")
			return err
		}
		defer rows.Close()

		list = make(storage.GaugeCollection, 0)
		for rows.Next() {
			var name string
			var val sql.NullFloat64
			err = rows.Scan(&name, &val)
			if err != nil {
				logger.Log.Error(err.Error())
				return err
			}

			if !val.Valid {
				logger.Log.Info("found null element")
				return fmt.Errorf("not found")
			}

			list[name] = storage.Gauge(val.Float64)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return list, nil
}

func (p Postgres) GetGaugeByName(name string) (storage.Gauge, error) {
	req := `SELECT value FROM gauge WHERE name=$1 LIMIT 1`

	var val sql.NullFloat64
	err := common.Retry(func() error {
		row := p.db.QueryRow(context.Background(), req, name)

		err := row.Scan(&val)
		if err != nil {
			return err
		}

		if !val.Valid {
			logger.Log.Info("found null element")
			return fmt.Errorf("not found")
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return storage.Gauge(val.Float64), nil
}

func (p Postgres) GetCounterByName(name string) (storage.Counter, error) {
	req := `SELECT value FROM counter WHERE name=$1 LIMIT 1`

	var val sql.NullInt64
	err := common.Retry(func() error {
		row := p.db.QueryRow(context.Background(), req, name)

		err := row.Scan(&val)
		if err != nil {
			return err
		}

		if !val.Valid {
			logger.Log.Error("error while validate counter value from db")
			return fmt.Errorf("error while validate counter value from db")
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return storage.Counter(val.Int64), nil
}

func (p Postgres) GetAll() (storage.Database, error) {
	gaugeList, err := p.GetAllGauge()
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
		return common.Retry(func() error {
			err := p.db.Ping(context.Background())
			if err != nil {
				return err
			}

			return nil
		})
	} else {
		return fmt.Errorf("db is not connected")
	}
}

func (p Postgres) Close() {
	p.db.Close()
}
