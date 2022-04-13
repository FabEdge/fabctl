package cert

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
	fclient "github.com/fabedge/fabedge/pkg/operator/client"
	certutil "github.com/fabedge/fabedge/pkg/util/cert"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newVerifyCmd(clientGetter types.ClientGetter) *cobra.Command {
	var commonOptions CommonOptions
	var selector string

	cmd := &cobra.Command{
		Use:   "verify [secretNames]",
		Short: "Verity your TLS secrets with specified CA",
		Example: `Verify specified TLS secrets:

	fabctl cert verify edge-tls edge2-tls

Verify TLS secrets by selectors:

	fabctl cert verify -l fabedge.io/created-by=fabctl

Verify TLS secrets using specified CA secret:

	fabctl cert verify --ca-secret=fabedge-ca

Verify TSL secrets using host cluster's API server:

	fabctl cert verify --remote --api-server-address=http://host-cluster/`,
		Run: func(cmd *cobra.Command, args []string) {
			cli := newClient(clientGetter)

			var caDER []byte
			if commonOptions.Remote() {
				cacert, err := fclient.GetCertificate(commonOptions.APIServerAddress)
				if err != nil {
					util.Exitf("failed to get CA certificate: %s\n", err)
				}
				caDER = cacert.DER
			} else {
				caDER, _ = cli.getCertAndKeyAsDER(commonOptions.CASecret)
			}

			verify := func(secret corev1.Secret) {
				certDER, _ := cli.getCertAndKeyFromSecret(secret)
				usages := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

				if err := certutil.VerifyCert(caDER, certDER, usages); err != nil {
					fmt.Fprintf(os.Stderr, "%s is invalid: %s\n", secret.Name, err)
				} else {
					fmt.Fprintf(os.Stdin, "%s is valid\n", secret.Name)
				}
			}

			if len(args) > 0 {
				for _, secretName := range args {
					secret := cli.getSecret(secretName)
					verify(secret)
				}

				return
			}

			var secretList corev1.SecretList
			l, err := labels.Parse(selector)
			util.CheckError(err)

			err = cli.List(context.Background(), &secretList, client.MatchingLabelsSelector{Selector: l})
			util.CheckError(err)

			for _, secret := range secretList.Items {
				verify(secret)
			}
		},
	}

	fs := cmd.Flags()
	commonOptions.AddFlags(fs)

	usage := "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Selectors will be ignored if you provide a secretName."
	fs.StringVarP(&selector, "selector", "l", "fabedge.io/created-by=fabedge-operator", usage)
	return cmd
}
