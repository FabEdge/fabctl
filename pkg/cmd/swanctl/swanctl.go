package swanctl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
)

func New(clientGetter types.ClientGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swanctl [command] edge [flags]",
		Short: "Execute swanctl command in strongswan containers",
		Long:  `Execute swanctl command in strongswan containers. There are four subcommands and you can also execute other swanctl subcommands`,
		Example: `
fabctl swanctl list-conns edge1

To execute command on connectors, just input:

fabctl swanctl list-conns connector

To execute others swanctl commands, input like this:
fabctl swanctl edge1 -- --version
fabctl swanctl connector -- --version
`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			execute(cli, args[0], args[1:]...)
		},
	}

	cmd.AddCommand(newSubCommand(
		"--list-conns",
		clientGetter,
		"list-conns edge [flags]",
		"List loaded configurations of strongswan container in specified edge",
		false, false,
	))
	cmd.AddCommand(newSubCommand(
		"--list-sa",
		clientGetter,
		"list-sa [edge] [flags]",
		"List currently active IKE_SAs of strongswan container in specified edge",
		true, false,
	))
	cmd.AddCommand(newSubCommand(
		"--initiate",
		clientGetter,
		"initiate edge [flags]",
		"Initiate connection of strongswan container in specified edge",
		true, true,
	))
	cmd.AddCommand(newSubCommand(
		"--terminate",
		clientGetter,
		"terminate edge [flags]",
		"Terminate connection of strongswan container in specified edge",
		true, true,
	))

	return cmd
}

func newSubCommand(name string, clientGetter types.ClientGetter, usage, short string, useIKE, useChild bool) *cobra.Command {
	flags := &swanctlFlags{}
	cmd := &cobra.Command{
		Use:   usage,
		Short: short,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			execute(cli, args[0], flags.build(name)...)
		},
	}

	flags.addRawAndPretty(cmd.Flags())
	if useIKE {
		flags.addIKE(cmd.Flags())
	}

	if useChild {
		flags.addChild(cmd.Flags())
	}

	return cmd
}

func execute(client *types.Client, edgeName string, flags ...string) {
	if edgeName == "connector" {
		executeOnConnectors(client, flags...)
		return
	}

	podName := edgeName
	if !strings.HasPrefix(edgeName, "fabedge-connector") {
		podName = fmt.Sprintf("fabedge-agent-%s", edgeName)
	}

	cmd := append([]string{"swanctl"}, flags...)

	fmt.Printf("========================== %s =================================\n", podName)
	err := client.Exec(podName, "strongswan", cmd)
	util.CheckError(err)
}

func executeOnConnectors(cli *types.Client, cmdFlags ...string) {
	var pods corev1.PodList
	err := cli.List(context.Background(), &pods, client.MatchingLabels{
		"app": "fabedge-connector",
	})
	util.CheckError(err)

	if len(pods.Items) == 0 {
		fmt.Fprintln(os.Stderr, "no connectors found")
	}

	for _, pod := range pods.Items {
		execute(cli, pod.Name, cmdFlags...)
	}
}
