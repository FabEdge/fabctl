package nodes

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
	nodeutil "github.com/fabedge/fabedge/pkg/util/node"
)

func New(clientGetter types.ClientGetter) *cobra.Command {
	var selector string
	var edgeOnly bool
	cmd := &cobra.Command{
		Use:   "nodes [node1] [node2]...",
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
				if edgeOnly {
					nodes, err = cli.ListNodes(context.Background(), cluster.EdgeLabels)
					util.CheckError(err)
				} else {
					l, err := labels.Parse(selector)
					util.CheckError(err)

					var nodeList corev1.NodeList
					err = cli.List(context.Background(), &nodeList, client.MatchingLabelsSelector{Selector: l})
					util.CheckError(err)

					nodes = nodeList.Items
				}
			}

			for _, node := range nodes {
				displayNodeInfo(node, cluster)
			}
		},
	}

	fs := cmd.Flags()

	usage := "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Selectors will be ignored if you provide any nodeName."
	fs.StringVarP(&selector, "selector", "l", "", usage)
	fs.BoolVarP(&edgeOnly, "edge-only", "e", false, "Display edge nodes only. If this flag is set to true, then selector won't work")
	return cmd
}

func displayNodeInfo(node corev1.Node, cluster *types.Cluster) {
	endpoint := cluster.NewEndpoint(node)
	podCIDRs := nodeutil.GetPodCIDRs(node)

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
EdgePodCIDRs:     %s
Communities:      %s
Peers:            %s
`,
		node.Name,
		strings.Join(endpoint.PublicAddresses, ","),
		strings.Join(endpoint.NodeSubnets, ","),
		strings.Join(podCIDRs, ","),
		strings.Join(endpoint.Subnets, ","),
		strings.Join(communityNames, ","),
		strings.Join(peers.List(), ","),
	)
}
