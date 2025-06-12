package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("incorrect")
	if err != nil {
		os.Exit(1) // want `alling os.Exit in main.main is not allowed`
	}

	content, err := io.ReadAll(file)
	if err != nil {
		os.Exit(1) // want `alling os.Exit in main.main is not allowed`
	}

	fmt.Println(string(content))
}
