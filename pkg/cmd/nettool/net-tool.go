package nettool

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
	"github.com/fabedge/fabedge/pkg/common/constants"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func New(clientGetter types.ClientGetter) *cobra.Command {
	var (
		podName        string
		image          string
		useHostNetwork bool
		httpPort       int32
		httpsPort      int32
	)

	cmd := &cobra.Command{
		Use:   "net-tool [command] nodeName [flags]",
		Short: "Create a net-tool pod on specified node for networking diagnosis purpose",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli, err := clientGetter.GetClient()
			util.CheckError(err)

			nodeName := args[0]
			netToolPod := newNetToolPod(nodeName, podName, cli.GetNamespace(), image, useHostNetwork, httpPort, httpsPort)

			if err = cli.Create(context.Background(), &netToolPod); err != nil {
				if errors.IsAlreadyExists(err) {
					fmt.Printf("Pod %s/%s is already existing\n", netToolPod.Namespace, netToolPod.Name)
					return
				} else {
					util.CheckError(err)
				}
			}
			fmt.Printf("Pod %s/%s is created\n", netToolPod.Namespace, netToolPod.Name)
		},
	}

	fs := cmd.Flags()
	fs.StringVarP(&image, "image", "i", "fabedge/net-tool:v0.1.0", "The image of net-tool pod")
	fs.StringVar(&podName, "podName", "", "The podName of generated pod, if this value is empty, fabctl will use podName derived from node podName")
	fs.BoolVar(&useHostNetwork, "host", false, "Use host network or not")
	fs.Int32Var(&httpPort, "http-port", 30080, "The default http port for net-tool pod")
	fs.Int32Var(&httpsPort, "https-port", 30443, "The default https port for net-tool pod")

	return cmd
}

func newNetToolPod(nodeName, podName, namespace, image string, useHostNetwork bool, httpPort, httpsPort int32) corev1.Pod {
	if podName == "" {
		if useHostNetwork {
			podName = fmt.Sprintf("host-net-tool-%s", nodeName)
		} else {
			podName = fmt.Sprintf("net-tool-%s", nodeName)
		}
	}

	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				constants.KeyFabedgeAPP: "net-tool",
				constants.KeyCreatedBy:  "fabctl",
			},
		},
		// change default port to avoid ports conflict with host service's endpoints
		Spec: corev1.PodSpec{
			HostNetwork:                  useHostNetwork,
			NodeName:                     nodeName,
			DNSPolicy:                    corev1.DNSClusterFirstWithHostNet,
			AutomountServiceAccountToken: new(bool),
			Containers: []corev1.Container{
				{
					Name:            "net-tool",
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: httpPort,
						},
						{
							Name:          "https",
							ContainerPort: httpsPort,
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "HTTP_PORT",
							Value: fmt.Sprint(httpPort),
						},
						{
							Name:  "HTTPS_PORT",
							Value: fmt.Sprint(httpsPort),
						},
					},
				},
			},
			Tolerations: []corev1.Toleration{
				{
					Key:      "",
					Operator: corev1.TolerationOpExists,
				},
			},
		},
	}
}
