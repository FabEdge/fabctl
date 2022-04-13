package cmd

import (
	"github.com/spf13/cobra"

	"github.com/fabedge/fabctl/pkg/cmd/cert"
	"github.com/fabedge/fabctl/pkg/cmd/clusterinfo"
	"github.com/fabedge/fabctl/pkg/cmd/edges"
	"github.com/fabedge/fabctl/pkg/cmd/images"
	"github.com/fabedge/fabctl/pkg/cmd/ping"
	"github.com/fabedge/fabctl/pkg/cmd/swanctl"
	"github.com/fabedge/fabctl/pkg/cmd/topology"
	"github.com/fabedge/fabctl/pkg/cmd/version"
	"github.com/fabedge/fabctl/pkg/types"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fabctl <command> <subcommand>",
	}

	clientFactory := types.NewClientFlags()
	clientFactory.AddFlags(cmd.PersistentFlags())

	cmd.AddCommand(clusterinfo.New(clientFactory))
	cmd.AddCommand(ping.New(clientFactory))
	cmd.AddCommand(images.New(clientFactory))
	cmd.AddCommand(edges.New(clientFactory))
	cmd.AddCommand(swanctl.New(clientFactory))
	cmd.AddCommand(topology.New(clientFactory))
	cmd.AddCommand(cert.New(clientFactory))
	cmd.AddCommand(version.New())

	return cmd
}
