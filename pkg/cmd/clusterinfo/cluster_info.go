package clusterinfo

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
)

type Cluster struct {
	client *types.Client
	Name   string
	Role   string

	CNIType                string
	EdgePodCIDR            string
	ConnectorPublicAddress string
	ConnectorSubnets       string

	Zone   string
	Region string
}

func New(clientGetter types.ClientGetter) *cobra.Command {
	return &cobra.Command{
		Use:   "cluster-info",
		Short: "Show information related to FabEdge of a cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			cluster := Cluster{client: cli}
			cluster.extractValuesFromOperator()
			cluster.extractTopology()

			fmt.Printf(`
Name:                       %s
Role:                       %s
Region:                     %s
Zone:                       %s
CNI Type:                   %s
EdgePodCIDR:                %s
Connector Public Addresses: %s
Connector Subnets:          %s
`,
				cluster.Name, cluster.Role,
				cluster.Region, cluster.Zone,
				cluster.CNIType, cluster.EdgePodCIDR,
				cluster.ConnectorPublicAddress, cluster.ConnectorSubnets,
			)
		},
	}
}

func (c *Cluster) extractValuesFromOperator() {
	operator, err := c.client.GetDeployment(context.Background(), "fabedge-operator")
	if err != nil {
		util.Exitf("failed to get fabedge-operator deployment: %s\n", err)
	}

	args := types.NewArgs(operator.Spec.Template.Spec.Containers[0].Args)
	c.Name = args.GetValue("cluster")
	c.Role = args.GetValue("cluster-role")
	c.CNIType = args.GetValue("cni-type")
	c.EdgePodCIDR = args.GetValue("edge-pod-cidr")
	c.ConnectorPublicAddress = args.GetValue("connector-public-addresses")
	c.ConnectorSubnets = args.GetValue("connector-subnets")
}

func (c *Cluster) extractTopology() {
	serviceHub, err := c.client.GetDeployment(context.Background(), "service-hub")
	switch {
	case err == nil:
		args := types.NewArgs(serviceHub.Spec.Template.Spec.Containers[0].Args)
		c.Region = args.GetValue("region")
		c.Zone = args.GetValue("zone")
	case errors.IsNotFound(err):
		util.Exitf("service-hub deployment is not found\n")
	default:
		util.Exitf("failed to get service-hub deployment: %s\n", err)
	}
}
