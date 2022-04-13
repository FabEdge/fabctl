package cert

import (
	"github.com/fabedge/fabctl/pkg/types"
	"github.com/spf13/cobra"
)

func New(clientGetter types.ClientGetter) *cobra.Command {
	rootCMD := &cobra.Command{
		Use:   "cert",
		Short: "A TLS certificate generator to facilitate deploying FabEdge",
	}

	rootCMD.AddCommand(newGenerateCmd(clientGetter))
	rootCMD.AddCommand(newViewCmd(clientGetter))
	rootCMD.AddCommand(newVerifyCmd(clientGetter))
	return rootCMD
}
