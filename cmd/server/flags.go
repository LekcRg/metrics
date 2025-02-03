package main

import "flag"

var addrFlag string

func parseFlags() {
	flag.StringVar(&addrFlag, "a", "localhost:8080", "address for run server")

	flag.Parse()
}
