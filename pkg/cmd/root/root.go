package root

import (
	"github.com/spf13/cobra"

	"github.com/fabedge/fabctl/pkg/cmd/version"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fabctl <command> <subcommand> [flags]",
	}

	cmd.AddCommand(version.New())

	return cmd
}
