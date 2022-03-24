package main

import (
	"fmt"
	"os"

	apis "github.com/fabedge/fabedge/pkg/apis/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/fabedge/fabctl/pkg/cmd"
)

func init() {
	_ = apis.AddToScheme(scheme.Scheme)
}

func main() {
	cmd := cmd.New()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
