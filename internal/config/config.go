package config

import (
	"flag"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/caarlos0/env/v11"
)

const defaultAddr = "localhost:8080"
const defaultLogLvl = "debug"
const defaultIsDev = false
const defaultStoreInterval = 300
const defaultFileStoragePath = "store.json"
const defaultRestore = false
const defaultReportInterval = 10
const defaultPollInterval = 2
const defaultHTTPS = false
const defaultDatabaseDSN = ""

// const defaultDatabaseDSN = "postgresql://postgres:postgres@localhost:5432/metrics"

type CommonConfig struct {
	Addr   string `env:"ADDRESS"`
	LogLvl string `env:"LOG_LVL"`
	IsDev  bool   `env:"IS_DEV"`
}

type ServerConfig struct {
	CommonConfig
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"-1"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	SyncSave        bool
}

type AgentConfig struct {
	CommonConfig
	ReportInterval int  `env:"REPORT_INTERVAL"`
	PollInterval   int  `env:"POLL_INTERVAL"`
	IsHTTPS        bool `env:"IS_HTTPS"`
}

func loadCommonCfg(cfg *CommonConfig) error {
	flag.StringVar(&cfg.Addr, "a", defaultAddr, "address for run server")
	flag.StringVar(&cfg.LogLvl, "l", defaultLogLvl, "logging level")
	flag.BoolVar(&cfg.IsDev, "dev", defaultIsDev, "is development")
	flag.Parse()

	var envVars CommonConfig
	err := env.Parse(&envVars)
	if err != nil {
		return err
	}
	if envVars.Addr != "" {
		cfg.Addr = envVars.Addr
	}

	if envVars.LogLvl != "" {
		cfg.LogLvl = envVars.LogLvl
	}

	if envVars.IsDev {
		cfg.IsDev = envVars.IsDev
	}

	return nil
}
func LoadServerCfg() ServerConfig {
	var cfg = ServerConfig{}

	flag.IntVar(&cfg.StoreInterval, "i", defaultStoreInterval, "time is seconds to save db to store(file)")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "path to save store")
	flag.BoolVar(&cfg.Restore, "r", defaultRestore, "restore db from file")
	flag.StringVar(&cfg.DatabaseDSN, "d", defaultDatabaseDSN, "Postgres database DSN")
	err := loadCommonCfg(&cfg.CommonConfig)
	if err != nil {
		logger.Log.Error("Error while load common config")
	}
	// flag.Parse() in loadCommonCfg()

	var envVars ServerConfig
	err = env.Parse(&envVars)
	if err != nil {
		logger.Log.Error("Error env vars")
	}

	if envVars.StoreInterval >= 0 {
		cfg.StoreInterval = envVars.StoreInterval
	}

	if envVars.FileStoragePath != "" {
		cfg.FileStoragePath = envVars.FileStoragePath
	}

	if envVars.Restore {
		cfg.Restore = envVars.Restore
	}

	if envVars.DatabaseDSN != "" {
		cfg.DatabaseDSN = envVars.DatabaseDSN
	}

	if cfg.DatabaseDSN != "" {
		cfg.Restore = false
	} else {
		cfg.SyncSave = cfg.StoreInterval == 0
	}

	return cfg
}

func LoadAgentCfg() AgentConfig {
	cfg := AgentConfig{}

	flag.IntVar(&cfg.ReportInterval, "r", defaultReportInterval, "interval for sending runtime metrics")
	flag.IntVar(&cfg.PollInterval, "p", defaultPollInterval, "interval for getting runtime metrics")
	flag.BoolVar(&cfg.IsHTTPS, "s", defaultHTTPS, "https true/false, default false")
	err := loadCommonCfg(&cfg.CommonConfig)
	if err != nil {
		logger.Log.Error("Error while load common config")
	}
	// flag.Parse() in loadCommonCfg()

	var envVars AgentConfig
	err = env.Parse(&envVars)
	if err != nil {
		logger.Log.Error("Error parse flags")
	}

	if envVars.ReportInterval != 0 {
		cfg.ReportInterval = envVars.ReportInterval
	}

	if envVars.PollInterval != 0 {
		cfg.PollInterval = envVars.PollInterval
	}

	if envVars.IsHTTPS {
		cfg.IsHTTPS = envVars.IsHTTPS
	}

	return cfg
}
