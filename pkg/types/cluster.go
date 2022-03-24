package types

import (
	"context"
	"strings"

	apisv1 "github.com/fabedge/fabedge/pkg/apis/v1alpha1"
	"github.com/fabedge/fabedge/pkg/common/constants"
	ftypes "github.com/fabedge/fabedge/pkg/operator/types"
	nodeutil "github.com/fabedge/fabedge/pkg/util/node"
)

type Cluster struct {
	client            *Client
	Name              string
	CNIType           string
	EdgeLabels        map[string]string
	EndpointIDFormat  string
	NewEndpoint       ftypes.NewEndpointFunc
	EdgeToCommunities map[string][]string
	Communities       map[string]apisv1.Community
}

func NewCluster(client *Client) *Cluster {
	return &Cluster{
		client:            client,
		Communities:       make(map[string]apisv1.Community),
		EdgeToCommunities: make(map[string][]string),
	}
}

func (cluster *Cluster) ExtractArgumentsFromFabEdge() error {
	operator, err := cluster.client.GetDeployment(context.Background(), "fabedge-operator")
	if err != nil {
		return err
	}

	args := NewArgs(operator.Spec.Template.Spec.Containers[0].Args)
	cluster.Name = args.GetValue("cluster")
	cluster.CNIType = args.GetValue("cni-type")
	cluster.EndpointIDFormat = args.GetValueOrDefault("endpoint-id-format", "C=CN, O=fabedge.io, CN={node}")
	cluster.EdgeLabels = parseLabels(args.GetValueOrDefault("edge-labels", "C=CN, O=fabedge.io, CN={node}"))

	var getPodCIDR ftypes.PodCIDRsGetter
	switch cluster.CNIType {
	case constants.CNICalico:
		getPodCIDR = nodeutil.GetPodCIDRsFromAnnotation
	case constants.CNIFlannel:
		getPodCIDR = nodeutil.GetPodCIDRs
	}

	_, _, cluster.NewEndpoint = ftypes.NewEndpointFuncs(cluster.Name, cluster.EndpointIDFormat, getPodCIDR)

	return nil
}

func (cluster *Cluster) LoadCommunities() error {
	var communityList apisv1.CommunityList
	err := cluster.client.List(context.Background(), &communityList)
	if err != nil {
		return err
	}

	for _, community := range communityList.Items {
		cluster.Communities[community.Name] = community
		for _, epName := range community.Spec.Members {
			cluster.EdgeToCommunities[epName] = append(cluster.EdgeToCommunities[epName], community.Name)
		}
	}

	return nil
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
