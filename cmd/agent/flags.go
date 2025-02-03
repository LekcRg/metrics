package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

var addrFlag string
var reportInterval int
var pollInterval int
var https bool

type config struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&addrFlag, "a", "localhost:8080", "server address")
	flag.IntVar(&reportInterval, "r", 10, "interval for sending runtime metrics")
	flag.IntVar(&pollInterval, "p", 2, "interval for getting runtime metrics")
	flag.BoolVar(&https, "s", false, "https true/false, default false")

	flag.Parse()

	var cfg config
	err := env.Parse(&cfg)

	if err != nil {
		fmt.Println("Error parse env")
	}

	if cfg.Addr != "" {
		addrFlag = cfg.Addr
	}

	if cfg.ReportInterval != 0 {
		reportInterval = cfg.ReportInterval
	}

	if cfg.PollInterval != 0 {
		pollInterval = cfg.ReportInterval
	}
}
