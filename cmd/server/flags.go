package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

var addrFlag string
var logLvl string
var isDev bool
var storeInterval int
var fileStoragePath string
var restore bool

type config struct {
	Addr            string `env:"ADDRESS"`
	LogLvl          string `env:"LOG_LVL"`
	IsDev           bool   `env:"IS_DEV"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

const defaultAddr = "localhost:8080"
const defaultLogLvl = "debug"
const defaultIsDev = false
const defaultStoreInterval = 300
const defaultFileStoragePath = "store.json"
const defaultRestore = false

func parseFlags() {
	flag.StringVar(&addrFlag, "a", defaultAddr, "address for run server")
	flag.StringVar(&logLvl, "l", defaultLogLvl, "logging level")
	flag.BoolVar(&isDev, "dev", defaultIsDev, "is development")
	flag.IntVar(&storeInterval, "i", defaultStoreInterval, "time is seconds to save db to store(file)")
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "path to save store")
	flag.BoolVar(&restore, "r", defaultRestore, "restore db from file")

	var cfg config
	err := env.Parse(&cfg)

	if err != nil {
		fmt.Println("Error parse env")
	}

	if cfg.Addr != "" {
		addrFlag = cfg.Addr
	}

	if cfg.LogLvl != "" {
		logLvl = cfg.LogLvl
	}

	if cfg.StoreInterval != 0 {
		storeInterval = cfg.StoreInterval
	}

	if cfg.FileStoragePath != "" {
		fileStoragePath = cfg.FileStoragePath
	}

	if cfg.Restore {
		restore = cfg.Restore
	}

	fmt.Printf(
		"Flags:\n    addrFlag: %s\n    logLvl: %v\n    isDev: %v\n    storeInterval: %v\n    fileStoragePath: %s\n    restore: %v\n",
		addrFlag, logLvl, isDev, storeInterval, fileStoragePath, restore,
	)
}
