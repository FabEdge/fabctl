package types

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
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
