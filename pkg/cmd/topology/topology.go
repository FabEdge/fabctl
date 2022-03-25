package topology

import (
	"context"
	"fmt"
	"os"
	"strings"

	apisv1 "github.com/fabedge/fabedge/pkg/apis/v1alpha1"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
)

func New(clientGetter types.ClientGetter) *cobra.Command {
	var output string
	var layout string

	cmd := &cobra.Command{
		Use:   "topology [filename] [flags]",
		Short: "Show the network topology of current cluster",
		Example: `
fabctl topology network.svg
fabctl topology -l dot -o dot network.dot 
`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			cluster := types.NewCluster(cli)
			util.CheckError(cluster.ExtractArgumentsFromFabEdge())
			util.CheckError(cluster.LoadCommunities())

			edgeNodes, err := cli.ListEdgeNodes(context.Background(), cluster.EdgeLabels)
			util.CheckError(err)

			clusters, err := cli.ListClusters(context.Background())
			util.CheckError(err)

			endpoints := make(map[string]Endpoint)
			for _, c := range clusters {
				for _, ep := range c.Spec.EndPoints {
					endpoints[ep.Name] = Endpoint{
						Endpoint:    ep,
						ClusterName: c.Name,
						Peers:       sets.NewString(),
						External:    c.Name != cluster.Name,
					}
				}
			}

			for _, node := range edgeNodes {
				ep := cluster.NewEndpoint(node)
				endpoints[ep.Name] = Endpoint{
					Endpoint:    ep,
					ClusterName: cluster.Name,
					Peers:       sets.NewString(),
				}
			}

			filename := ""
			if len(args) == 1 {
				filename = args[0]
			}

			renderTopology(cluster, endpoints, filename, graphviz.Format(output), layout)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", string(graphviz.SVG), "Output format, possible values: dot, svg, png, jpg.")
	cmd.Flags().StringVarP(&layout, "layout", "l", "sfdp", "Topology layout, check out https://graphviz.org/docs/layouts/ for possible options.")
	return cmd
}

type Endpoint struct {
	apisv1.Endpoint
	ClusterName string
	External    bool

	Peers sets.String

	GraphNode *cgraph.Node
}

func (e *Endpoint) BuildGraphNode(graph *cgraph.Graph) error {
	node, err := graph.CreateNode(e.Name)
	if err != nil {
		return err
	}

	node.SetStyle(cgraph.FilledNodeStyle)
	node.SetTooltip(fmt.Sprintf(`
Name: %s
PodCIDRs: %s
Node Subnets: %s
Public Addresses: %s
`,
		e.Name,
		strings.Join(e.Subnets, ","),
		strings.Join(e.NodeSubnets, ","),
		strings.Join(e.PublicAddresses, ","),
	))
	switch e.Type {
	case apisv1.Connector:
		if e.External {
			node.SetFillColor("#9acae1")
		} else {
			node.SetFillColor("forestgreen")
		}
	case apisv1.EdgeNode:
		if e.External {
			node.SetFillColor("#deebf7")
		} else {
			node.SetFillColor("darkseagreen3")
		}
	}
	e.GraphNode = node

	return err
}

func renderTopology(cluster *types.Cluster, endpoints map[string]Endpoint, filename string, format graphviz.Format, layout string) {
	g := graphviz.New()
	graph, err := g.Graph()
	util.CheckError(err)

	graph.SetLayout(layout)

	for name, ep := range endpoints {
		util.CheckError((&ep).BuildGraphNode(graph))
		endpoints[name] = ep
	}

	// draw lines between connector and edge endpoints of current cluster
	connectorName := fmt.Sprintf("%s.connector", cluster.Name)
	connector := endpoints[connectorName]
	for _, endpoint := range endpoints {
		if endpoint.ClusterName == cluster.Name {
			createEdgeBetweenEndpoints(connector, endpoint, graph)
		}
	}

	// draw lines between members of communities
	for _, community := range cluster.Communities {
		for _, epName := range community.Spec.Members {
			endpoint, ok := endpoints[epName]
			if !ok {
				continue
			}

			for _, peerName := range community.Spec.Members {
				if peer, ok := endpoints[peerName]; ok {
					createEdgeBetweenEndpoints(endpoint, peer, graph)
				}
			}
		}
	}

	if filename == "" {
		err = g.Render(graph, format, os.Stdout)
		util.CheckError(err)
	} else {
		err = g.RenderFilename(graph, format, filename)
		util.CheckError(err)

		fmt.Printf("Topology information is written to %s.\n", filename)
		if format != graphviz.XDOT {
			fmt.Println("If you execute `fabctl topology` on a remote computer, it is recommended to open a http server to view the picture, e.g.: python -m http.server 8080.")
		}
	}

}

func createEdgeBetweenEndpoints(e1, e2 Endpoint, graph *cgraph.Graph) {
	if e1.Name == e2.Name {
		return
	}

	if e1.Peers.Has(e2.Name) || e2.Peers.Has(e1.Name) {
		return
	}

	edgeName := fmt.Sprintf("%s-%s", e1.Name, e2.Name)
	edge, err := graph.CreateEdge(edgeName, e1.GraphNode, e2.GraphNode)

	util.CheckError(err)
	edge.SetArrowHead(cgraph.NoneArrow)
	edge.SetArrowTail(cgraph.NoneArrow)

	e1.Peers.Insert(e2.Name)
	e2.Peers.Insert(e1.Name)
}
