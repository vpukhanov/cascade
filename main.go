package main

import (
	"os"

	"github.com/vpukhanov/cascade/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
