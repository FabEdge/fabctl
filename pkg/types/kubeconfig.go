package types

import (
	"flag"

	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ClientGetter interface {
	GetConfig() (*rest.Config, error)
	GetClient() (*Client, error)
}

type ClientFactory struct {
	Namespace string
}

func NewClientFlags() *ClientFactory {
	return &ClientFactory{}
}

func (cfg *ClientFactory) AddFlags(fs *pflag.FlagSet) {
	fs.AddGoFlagSet(flag.CommandLine)
	fs.StringVarP(&cfg.Namespace, "namespace", "n", "fabedge", "The namespace where FabEdge is deployed.")
}

func (cfg ClientFactory) GetConfig() (*rest.Config, error) {
	return config.GetConfig()
}

func (cfg ClientFactory) GetClient() (*Client, error) {
	restConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	cli, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		config:    restConfig,
		Client:    cli,
		clientset: clientset,
		namespace: cfg.Namespace,
	}, nil
}
