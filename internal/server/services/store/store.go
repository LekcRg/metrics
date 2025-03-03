package store

import (
	"encoding/json"
	"os"
	"time"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type db interface {
	GetAll() (storage.Database, error)
	SaveManyGauge(storage.GaugeCollection) error
	SaveManyCounter(storage.CounterCollection) error
}

type Store struct {
	cfg config.ServerConfig
	db  db
}

func NewStore(storage db, cfg config.ServerConfig) *Store {
	return &Store{
		cfg: cfg,
		db:  storage,
	}
}

func (s Store) Save() error {
	metrics, err := s.db.GetAll()
	if err != nil {
		logger.Log.Error("Error while getting all metrics from storage")
		return err
	}

	storeJSON, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("Error while marshal json store")
		return err
	}

	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Fatal(err.Error())
		return err
	}

	_, err = file.Write(storeJSON)
	if err != nil {
		logger.Log.Error("Error while saving file")
		return err
	}
	file.Close()

	logger.Log.Info("Success save store to file " + s.cfg.FileStoragePath)
	return nil
}

func (s Store) StartSaving() {
	ticker := time.NewTicker(time.Duration(s.cfg.StoreInterval) * time.Second)
	for range ticker.C {
		err := s.Save()
		if err != nil {
			logger.Log.Error("Error while save")
		}
	}
}

func (s Store) Restore() error {
	file, err := os.ReadFile(s.cfg.FileStoragePath)
	if err != nil {
		logger.Log.Error("Can't open file")
		return err
	}
	var storage storage.Database
	err = json.Unmarshal(file, &storage)
	if err != nil {
		return err
	}

	if len(storage.Gauge) > 0 {
		err := s.db.SaveManyGauge(storage.Gauge)
		if err != nil {
			return err
		}
	}

	if len(storage.Counter) > 0 {
		err := s.db.SaveManyCounter(storage.Counter)
		if err != nil {
			return err
		}
	}

	logger.Log.Info("Success restore data from file " + s.cfg.FileStoragePath)

	return nil
}
