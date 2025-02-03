package main

import "flag"

var addrFlag string
var reportInterval int
var pollInterval int
var https bool

func parseFlags() {
	flag.StringVar(&addrFlag, "a", "localhost:8080", "server address")
	flag.IntVar(&reportInterval, "r", 10, "interval for sending runtime metrics")
	flag.IntVar(&pollInterval, "p", 2, "interval for getting runtime metrics")
	flag.BoolVar(&https, "s", false, "https true/false, default false")

	flag.Parse()
}
