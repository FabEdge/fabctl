package root

import (
	"github.com/spf13/cobra"

	"github.com/fabedge/fabctl/pkg/cmd/clusterinfo"
	"github.com/fabedge/fabctl/pkg/cmd/version"
	"github.com/fabedge/fabctl/pkg/types"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fabctl <command> <subcommand> [flags]",
	}

	kubeConfig := types.NewKubeConfig()
	kubeConfig.AddFlags(cmd.PersistentFlags())

	cmd.AddCommand(clusterinfo.New(kubeConfig))
	cmd.AddCommand(version.New())

	return cmd
}
