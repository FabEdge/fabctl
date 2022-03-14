package version

import (
	"github.com/spf13/cobra"

	"github.com/fabedge/fabctl/pkg/about"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			about.DisplayVersion()
		},
	}
}
