package types

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	client.Client
	config *rest.Config
}

func (c Client) GetDeployment(ctx context.Context, namespace, name string) (appsv1.Deployment, error) {
	var deploy appsv1.Deployment
	err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &deploy)

	return deploy, err
}

func (c Client) GetDaemonSet(ctx context.Context, namespace, name string) (appsv1.DaemonSet, error) {
	var ds appsv1.DaemonSet
	err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &ds)
	return ds, err
}

func (c Client) GetNode(ctx context.Context, name string) (corev1.Node, error) {
	var node corev1.Node
	err := c.Get(ctx, client.ObjectKey{Name: name}, &node)
	return node, err
}
