package main

import (
	"os"

	"github.com/fabedge/fabctl/pkg/cmd/root"
)

func main() {
	cmd := root.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
