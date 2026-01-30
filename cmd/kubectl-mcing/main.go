package main

import (
	"os"

	"github.com/kmdkuk/mcing/cmd/kubectl-mcing/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
