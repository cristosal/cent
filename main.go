package main

import (
	"log"

	"github.com/cristosal/cent/cmd"
)

func main() {
	log.Fatal(cmd.Root.Execute())
}
