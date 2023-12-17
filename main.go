package main

import (
	"os"

	"github.com/cristosal/cent/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
