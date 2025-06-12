package notmainpackage

import (
	"fmt"
	"io"
	"os"
)

func main() {
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
