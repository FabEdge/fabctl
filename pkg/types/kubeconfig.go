package types

import (
	"flag"
	"k8s.io/client-go/kubernetes"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ClientGetter interface {
	GetConfig() (*rest.Config, error)
	GetClient() (*Client, error)
}

type KubeConfig struct{}

func NewKubeConfig() *KubeConfig {
	return &KubeConfig{}
}

func (cfg KubeConfig) AddFlags(fs *pflag.FlagSet) {
	fs.AddGoFlagSet(flag.CommandLine)
}

func (cfg KubeConfig) GetConfig() (*rest.Config, error) {
	return config.GetConfig()
}

func (cfg KubeConfig) GetClient() (*Client, error) {
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

	return &Client{config: restConfig, Client: cli, clientset: clientset}, nil
}
