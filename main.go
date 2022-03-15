package main

import (
	"fmt"
	"os"

	"github.com/fabedge/fabctl/pkg/cmd/root"
)

func main() {
	cmd := root.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
