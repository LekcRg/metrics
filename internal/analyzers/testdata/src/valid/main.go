package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("incorrect")
	if err != nil {
		panic("file does not exist")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		panic("error while reading the file")
	}

	fmt.Println(string(content))
}
