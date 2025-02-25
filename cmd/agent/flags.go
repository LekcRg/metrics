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

const defaultAddr = "localhost:8080"
const defaultReportInterval = 10
const defaultPollInterval = 2
const defaultHTTPS = false

func parseFlags() {
	flag.StringVar(&addrFlag, "a", defaultAddr, "server address")
	flag.IntVar(&reportInterval, "r", defaultReportInterval, "interval for sending runtime metrics")
	flag.IntVar(&pollInterval, "p", defaultPollInterval, "interval for getting runtime metrics")
	flag.BoolVar(&https, "s", defaultHTTPS, "https true/false, default false")

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
		pollInterval = cfg.PollInterval
	}
}
