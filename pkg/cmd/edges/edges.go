package edges

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func New(clientGetter types.ClientGetter) *cobra.Command {
	return &cobra.Command{
		Use:   "edges [node1] [node2]...",
		Short: "Show network information about edge nodes",
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			cluster := types.NewCluster(cli)
			util.CheckError(cluster.ExtractArgumentsFromFabEdge())
			util.CheckError(cluster.LoadCommunities())

			var nodes []corev1.Node
			if len(args) > 0 {
				for _, name := range args {
					node, err := cli.GetNode(context.Background(), name)
					if err != nil {
						fmt.Fprint(os.Stderr, err.Error())
					} else {
						nodes = append(nodes, node)
					}
				}
			} else {
				nodes, err = cli.ListEdgeNodes(context.Background(), cluster.EdgeLabels)
				util.CheckError(err)
			}

			for _, node := range nodes {
				displayNodeInfo(node, cluster)
			}
		},
	}
}

func displayNodeInfo(node corev1.Node, cluster *types.Cluster) {
	endpoint := cluster.NewEndpoint(node)

	communityNames, peers := cluster.EdgeToCommunities[endpoint.Name], sets.NewString()
	for _, name := range communityNames {
		peers.Insert(cluster.Communities[name].Spec.Members...)
	}
	peers.Delete(endpoint.Name)

	fmt.Printf(`
Name:             %s
Public Addresses: %s
Node Subnets:     %s
PodCIDRs:         %s
Communities:      %s
Peers:            %s
`,
		node.Name,
		strings.Join(endpoint.PublicAddresses, ","),
		strings.Join(endpoint.NodeSubnets, ","),
		strings.Join(endpoint.Subnets, ","),
		strings.Join(communityNames, ","),
		strings.Join(peers.List(), ","),
	)
}
