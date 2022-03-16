package edges

import (
	"context"
	"fmt"
	"os"
	"strings"

	apisv1 "github.com/fabedge/fabedge/pkg/apis/v1alpha1"
	"github.com/fabedge/fabedge/pkg/common/constants"
	ftypes "github.com/fabedge/fabedge/pkg/operator/types"
	nodeutil "github.com/fabedge/fabedge/pkg/util/node"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
)

type Cluster struct {
	Name              string
	CNIType           string
	EdgeLabels        map[string]string
	EndpointIDFormat  string
	NewEndpoint       ftypes.NewEndpointFunc
	EdgeToCommunities map[string][]string
}

func New(clientGetter types.ClientGetter) *cobra.Command {
	return &cobra.Command{
		Use:   "edges [node1] [node2]...",
		Short: "Show network information about edge nodes",
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			cluster, err := getCluster(cli)
			util.CheckError(err)

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
				var nodeList corev1.NodeList
				err = cli.List(context.Background(), &nodeList, client.MatchingLabels(cluster.EdgeLabels))
				util.CheckError(err)
				nodes = nodeList.Items
			}

			for _, node := range nodes {
				displayNodeInfo(node, cluster)
			}
		},
	}
}

func getCluster(cli *types.Client) (Cluster, error) {
	operator, err := cli.GetDeployment(context.Background(), "fabedge", "fabedge-operator")
	if err != nil {
		return Cluster{}, err
	}

	var communityList apisv1.CommunityList
	err = cli.List(context.Background(), &communityList)
	if err != nil {
		return Cluster{}, err
	}

	args := types.NewArgs(operator.Spec.Template.Spec.Containers[0].Args)
	cluster := Cluster{
		Name:              args.GetValue("cluster"),
		CNIType:           args.GetValue("cni-type"),
		EndpointIDFormat:  args.GetValueOrDefault("endpoint-id-format", "C=CN, O=fabedge.io, CN={node}"),
		EdgeToCommunities: make(map[string][]string),
		EdgeLabels:        parseLabels(args.GetValueOrDefault("edge-labels", "C=CN, O=fabedge.io, CN={node}")),
	}

	var getPodCIDR ftypes.PodCIDRsGetter
	switch cluster.CNIType {
	case constants.CNICalico:
		getPodCIDR = nodeutil.GetPodCIDRsFromAnnotation
	case constants.CNIFlannel:
		getPodCIDR = nodeutil.GetPodCIDRs
	}

	_, _, cluster.NewEndpoint = ftypes.NewEndpointFuncs(cluster.Name, cluster.EndpointIDFormat, getPodCIDR)

	for _, community := range communityList.Items {
		for _, epName := range community.Spec.Members {
			cluster.EdgeToCommunities[epName] = append(cluster.EdgeToCommunities[epName], community.Name)
		}
	}

	return cluster, nil
}

func displayNodeInfo(node corev1.Node, cluster Cluster) {
	endpoint := cluster.NewEndpoint(node)
	fmt.Printf(`
Name: %s
Public Addresses: %s
Node Subnets: %s
PodCIDRs: %s
Communities: %s
`,
		node.Name,
		strings.Join(endpoint.PublicAddresses, ","),
		strings.Join(endpoint.NodeSubnets, ","),
		strings.Join(endpoint.Subnets, ","),
		strings.Join(cluster.EdgeToCommunities[endpoint.Name], ","),
	)
}

func parseLabels(labels string) map[string]string {
	labels = strings.TrimSpace(labels)

	parsedEdgeLabels := make(map[string]string)
	for _, label := range strings.Split(labels, ",") {
		parts := strings.SplitN(label, "=", 1)
		switch len(parts) {
		case 1:
			parsedEdgeLabels[parts[0]] = ""
		case 2:
			parsedEdgeLabels[parts[0]] = parts[1]
		default:
		}
	}

	return parsedEdgeLabels
}
