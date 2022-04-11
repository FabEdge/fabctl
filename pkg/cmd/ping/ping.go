package ping

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
	"github.com/fabedge/fabedge/pkg/common/constants"
)

const containerName = "net-tool"

func New(clientGetter types.ClientGetter) *cobra.Command {
	var image string
	var prepareTimeout time.Duration
	var pingDeadline uint
	var pingCount uint
	var keepPods bool

	cmd := &cobra.Command{
		Use:   "ping nodeName nodeName",
		Short: "Test if network between two nodes if works",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if pingCount == 0 {
				util.Exitf("Ping count should be greater that zero\n")
			}

			client, err := clientGetter.GetClient()
			util.CheckError(err)

			// prepare net tool pods
			var pod1, pod2 corev1.Pod
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), prepareTimeout)
				defer cancel()

				nodeName1, nodeName2 := args[0], args[1]
				pod1 = getOrCreateNetToolPod(ctx, client, nodeName1, image, prepareTimeout)
				pod2 = getOrCreateNetToolPod(ctx, client, nodeName2, image, prepareTimeout)
			}()

			if !keepPods {
				// delete net tool pods
				defer func() {
					ctx, cancel := context.WithTimeout(context.Background(), prepareTimeout)
					defer cancel()

					deletePod(ctx, client, pod1)
					deletePod(ctx, client, pod2)
				}()
			}

			ping(client, pod1, pod2, pingDeadline, pingCount)
			ping(client, pod2, pod1, pingDeadline, pingCount)
		},
	}

	fs := cmd.Flags()
	fs.StringVarP(&image, "net-tool-image", "i", "praqma/network-multitool:minimal", "The image of net-tool pod")
	fs.DurationVar(&prepareTimeout, "prepare-timeout", 30*time.Second, "The length of time to prepare net-tool pods which are used to execute ping command")
	fs.UintVar(&pingDeadline, "ping-deadline", 0, "The deadline argument of ping command")
	fs.UintVar(&pingCount, "ping-count", 5, "The count argument of ping command")
	fs.BoolVarP(&keepPods, "keep", "k", false, "keep pods after test if finished")
	return cmd
}

func getOrCreateNetToolPod(ctx context.Context, client *types.Client, nodeName string, image string, timeout time.Duration) corev1.Pod {
	var (
		podName   = fmt.Sprintf("net-tool-%s", nodeName)
		namespace = client.GetNamespace()

		pod corev1.Pod
		key = types.ObjectKey{Name: podName, Namespace: namespace}
	)

	err := client.Get(ctx, key, &pod)
	switch {
	case err == nil:
	case !errors.IsNotFound(err):
		util.CheckError(err)
	default:
		pod = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: client.GetNamespace(),
				Labels: map[string]string{
					constants.KeyFabedgeAPP: "net-tool",
					constants.KeyCreatedBy:  "fabctl",
				},
			},
			Spec: corev1.PodSpec{
				HostNetwork: false,
				NodeName:    nodeName,
				DNSPolicy:   corev1.DNSClusterFirstWithHostNet,
				// workaround, or it will fail at edgecore
				AutomountServiceAccountToken: new(bool),
				Containers: []corev1.Container{
					{
						Name:            containerName,
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
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

		util.CheckError(client.Create(ctx, &pod))
	}

	err = wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		err = client.Get(ctx, key, &pod)
		if err != nil {
			return false, err
		}

		return pod.Status.Phase == corev1.PodRunning, nil
	})
	util.CheckError(err)

	return pod
}

func ping(client *types.Client, pod1, pod2 corev1.Pod, deadline, count uint) {
	cmd := []string{"ping"}

	if deadline > 0 {
		cmd = append(cmd, "-w", fmt.Sprint(deadline))
	}

	if count == 0 {
		count = 1
	}
	cmd = append(cmd, "-c", fmt.Sprint(count))
	cmd = append(cmd, pod2.Status.PodIP)

	fmt.Printf("Ping from %s(%s) -> %s(%s) \n\n", pod1.Name, pod1.Status.PodIP, pod2.Name, pod2.Status.PodIP)
	util.CheckError(client.Exec(pod1.Name, containerName, cmd))
}

func deletePod(ctx context.Context, client *types.Client, pod corev1.Pod) {
	err := client.Delete(ctx, &pod)
	if err != nil && !errors.IsNotFound(err) {
		fmt.Fprintf(os.Stderr, "failed to delete pod %s/%s", pod.Namespace, pod.Namespace)
	}
}
