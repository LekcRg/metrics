package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/storage"
)

type Store struct {
	db  storage.Storage
	cfg config.ServerConfig
}

func NewStore(storage storage.Storage, cfg config.ServerConfig) *Store {
	return &Store{
		cfg: cfg,
		db:  storage,
	}
}

func (s Store) Save(ctx context.Context) error {
	metrics, err := s.db.GetAll(ctx)
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
		logger.Log.Error(err.Error())
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

func (s Store) StartSaving(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Duration(s.cfg.StoreInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stopped store auto saving")
			wg.Done()
			return
		case <-ticker.C:
			err := s.Save(ctx)
			if err != nil {
				logger.Log.Error("Error while save")
			}
		}
	}
}

func (s Store) Restore(ctx context.Context) error {
	if s.cfg.DatabaseDSN != "" {
		return fmt.Errorf("postgres doesn't support restore from file")
	}
	file, err := os.ReadFile(s.cfg.FileStoragePath)
	if err != nil {
		logger.Log.Error("Can't open file with path " + s.cfg.FileStoragePath)
		return err
	}
	var storage storage.Database
	err = json.Unmarshal(file, &storage)
	if err != nil {
		return err
	}

	err = s.db.UpdateMany(ctx, storage)
	if err != nil {
		logger.Log.Error("Error while restoring db")
		return err
	}

	logger.Log.Info("Success restore data from file " + s.cfg.FileStoragePath)

	return nil
}
