package images

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
)

type Images struct {
	client *types.Client

	Operator            string
	Agent               string
	AgentStrongSwan     string
	Connector           string
	ConnectorStrongSwan string
	CloudAgent          string
	ServiceHub          string
	FabDNS              string
}

func New(clientGetter types.ClientGetter) *cobra.Command {
	return &cobra.Command{
		Use:   "images",
		Short: "Show images of FabEdge and FabDNS",
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			images := Images{client: cli}
			images.extractImages()

			fmt.Printf(`
Operator:                 %s
Agent:                    %s
AgentStrongSwan:          %s
Connector:                %s
ConnectorStrongSwan:      %s
CloudAgent:               %s
ServiceHub:               %s
FabDNS:                   %s
`,
				images.Operator, images.Agent, images.AgentStrongSwan,
				images.Connector, images.ConnectorStrongSwan, images.CloudAgent,
				images.ServiceHub, images.FabDNS,
			)
		},
	}
}

func (images *Images) extractImages() {
	ctx := context.Background()

	operator, err := images.client.GetDeployment(ctx, "fabedge", "fabedge-operator")
	doIfNoError(err, func() {
		args := types.NewArgs(operator.Spec.Template.Spec.Containers[0].Args)
		images.Operator = operator.Spec.Template.Spec.Containers[0].Image
		images.Agent = args.GetValue("agent-image")
		images.AgentStrongSwan = args.GetValue("agent-strongswan-image")
	})

	connector, err := images.client.GetDeployment(ctx, "fabedge", "fabedge-connector")
	doIfNoError(err, func() {
		images.ConnectorStrongSwan = connector.Spec.Template.Spec.Containers[0].Image
		images.Connector = connector.Spec.Template.Spec.Containers[1].Image
	})

	cloudAgent, err := images.client.GetDaemonSet(ctx, "fabedge", "fabedge-cloud-agent")
	doIfNoError(err, func() {
		images.CloudAgent = cloudAgent.Spec.Template.Spec.Containers[0].Image
	})

	serviceHub, err := images.client.GetDeployment(ctx, "fabedge", "service-hub")
	doIfNoError(err, func() {
		images.ServiceHub = serviceHub.Spec.Template.Spec.Containers[0].Image
	})

	fabdns, err := images.client.GetDeployment(ctx, "fabedge", "fabdns")
	doIfNoError(err, func() {
		images.FabDNS = fabdns.Spec.Template.Spec.Containers[0].Image
	})
}

func doIfNoError(err error, fn func()) {
	if err == nil {
		fn()
	} else {
		fmt.Fprintf(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr)
	}
}
