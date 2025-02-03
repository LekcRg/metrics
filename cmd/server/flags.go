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

func parseFlags() {
	flag.StringVar(&addrFlag, "a", "localhost:8080", "address for run server")
	flag.Parse()

	var cfg config
	err := env.Parse(&cfg)

	if err != nil {
		fmt.Println("Error parse env")
	}

	if cfg.Addr != "" {
		addrFlag = cfg.Addr
	}

	fmt.Printf("addr: %s\n", addrFlag)
}
