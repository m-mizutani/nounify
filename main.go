package main

import (
	"os"

	"github.com/m-mizutani/nounify/pkg/controller/cli"
)

func main() {
	if cli.Run(os.Args) != nil {
		os.Exit(1)
	}
}
