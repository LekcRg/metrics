package buildinfo

import "fmt"

var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func Print() {
	fmt.Println("Build version: " + BuildVersion)
	fmt.Println("Build date: " + BuildDate)
	fmt.Println("Build commit: " + BuildCommit)
}
