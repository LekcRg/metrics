package main

import (
	"fmt"
	"io"
	"os"
)

var msg = "Hello from testdata!"

func main() {
	fmt.Println(msg)
}

func test() {
	file, err := os.Open("incorrect")
	if err != nil {
		os.Exit(1)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println(string(content))
}
