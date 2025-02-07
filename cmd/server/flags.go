package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

var addrFlag string

type config struct {
	Addr string `env:"ADDRESS"`
}

const defaultAddr = "localhost:8080"

func parseFlags() {
	flag.StringVar(&addrFlag, "a", defaultAddr, "address for run server")
	flag.Parse()

	var cfg config
	err := env.Parse(&cfg)

	if err != nil {
		fmt.Println("Error parse env")
	}

	if cfg.Addr != "" {
		addrFlag = cfg.Addr
	}
}
