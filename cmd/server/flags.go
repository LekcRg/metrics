package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

var addrFlag string
var logLvl string
var isDev bool

type config struct {
	Addr   string `env:"ADDRESS"`
	LogLvl string `env:"LOG_LVL"`
	IsDev  bool   `env:"IS_DEV"`
}

const defaultAddr = "localhost:8080"
const defaultLogLvl = "debug"
const defaultIsDev = false

func parseFlags() {
	flag.StringVar(&addrFlag, "a", defaultAddr, "address for run server")
	flag.StringVar(&logLvl, "l", defaultLogLvl, "logging level")
	flag.BoolVar(&isDev, "dev", defaultIsDev, "logging level")
	flag.Parse()

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

	if cfg.IsDev {
		isDev = cfg.IsDev
	}
}
