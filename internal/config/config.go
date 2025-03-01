package config

import (
	"flag"
	"fmt"

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

type commonConfig struct {
	Addr   string `env:"ADDRESS"`
	LogLvl string `env:"LOG_LVL"`
	IsDev  bool   `env:"IS_DEV"`
}

type serverConfig struct {
	commonConfig
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type agentConfig struct {
	commonConfig
	ReportInterval int  `env:"REPORT_INTERVAL"`
	PollInterval   int  `env:"POLL_INTERVAL"`
	IsHttps        bool `env:"IS_HTTPS"`
}

func loadCommonCfg(cfg *commonConfig) error {
	flag.StringVar(&cfg.Addr, "a", defaultAddr, "address for run server")
	flag.StringVar(&cfg.LogLvl, "l", defaultLogLvl, "logging level")
	flag.BoolVar(&cfg.IsDev, "dev", defaultIsDev, "is development")
	flag.Parse()

	var envVars commonConfig
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

func LoadAgentCfg() agentConfig {
	cfg := agentConfig{}

	flag.IntVar(&cfg.ReportInterval, "r", defaultReportInterval, "interval for sending runtime metrics")
	flag.IntVar(&cfg.PollInterval, "p", defaultPollInterval, "interval for getting runtime metrics")
	flag.BoolVar(&cfg.IsHttps, "s", defaultHTTPS, "https true/false, default false")
	err := loadCommonCfg(&cfg.commonConfig)
	if err != nil {
		logger.Log.Error("Error while load common config")
	}

	var envVars agentConfig
	err = env.Parse(&envVars)
	if err != nil {
		logger.Log.Error("Error parse flags")
	}
	// flag.Parse() in loadCommonCfg()

	if envVars.ReportInterval != 0 {
		cfg.ReportInterval = envVars.ReportInterval
	}

	if envVars.PollInterval != 0 {
		cfg.PollInterval = envVars.PollInterval
	}

	if envVars.IsHttps {
		cfg.IsHttps = envVars.IsHttps
	}

	cfgString := fmt.Sprintf("%+v\n", cfg)
	fmt.Println(cfgString)
	// logger.Log.Info(cfgString)

	return cfg
}

func LoadServerCfg() serverConfig {
	var cfg = serverConfig{}

	flag.IntVar(&cfg.StoreInterval, "i", defaultStoreInterval, "time is seconds to save db to store(file)")
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "path to save store")
	flag.BoolVar(&cfg.Restore, "r", defaultRestore, "restore db from file")
	loadCommonCfg(&cfg.commonConfig)
	err := loadCommonCfg(&cfg.commonConfig)
	if err != nil {
		logger.Log.Error("Error while load common config")
	}
	// flag.Parse() in loadCommonCfg()

	var envVars serverConfig
	err = env.Parse(&envVars)
	if err != nil {
		logger.Log.Error("Error env vars")
	}

	if envVars.StoreInterval != 0 {
		cfg.StoreInterval = envVars.StoreInterval
	}

	if envVars.FileStoragePath != "" {
		cfg.FileStoragePath = envVars.FileStoragePath
	}

	if envVars.Restore {
		cfg.Restore = envVars.Restore
	}

	cfgString := fmt.Sprintf("%+v\n", cfg)
	fmt.Println("cfg.IsDev")
	fmt.Println(cfg.IsDev)
	fmt.Println(cfgString)

	// logger.Log.Info(cfgString)

	return cfg
}
