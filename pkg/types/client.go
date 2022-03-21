package types

import (
	"context"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExecResult struct {
	Stdout string
	Stderr string
	Err    error
}

type Client struct {
	client.Client
	namespace string
	config    *rest.Config
	clientset kubernetes.Interface
}

func (c Client) GetDeployment(ctx context.Context, name string) (appsv1.Deployment, error) {
	var deploy appsv1.Deployment
	err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: c.namespace}, &deploy)

	return deploy, err
}

func (c Client) GetDaemonSet(ctx context.Context, name string) (appsv1.DaemonSet, error) {
	var ds appsv1.DaemonSet
	err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: c.namespace}, &ds)
	return ds, err
}

func (c Client) GetNode(ctx context.Context, name string) (corev1.Node, error) {
	var node corev1.Node
	err := c.Get(ctx, client.ObjectKey{Name: name}, &node)
	return node, err
}

func (c Client) Exec(podName, containerName string, cmd []string) error {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(c.namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})

	return err
}
