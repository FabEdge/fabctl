package cert

import (
	"crypto/x509"
	"net"

	fclient "github.com/fabedge/fabedge/pkg/operator/client"
	certutil "github.com/fabedge/fabedge/pkg/util/cert"
	"github.com/spf13/cobra"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
)

func newGenerateCmd(clientGetter types.ClientGetter) *cobra.Command {
	var commonOptions CommonOptions
	var certOptions CertOptions
	var saveOptions SaveOptions

	cmd := &cobra.Command{
		Use:   "gen commonName",
		Short: "Create a pair of certificate and private key using specified CA",
		Long:  `Create a pair of certificate and private key using specified CA. By default the certificate and key will be save to a secret named by your commonName, you can specify the secret name.`,
		Example: `Create a pair of certificate and private key with commonName "edge":

	fabctl cert gen edge

Create a pair of certificate and private key with commonName edge and save data to secret edge-tls:

	fabctl gen edge --secret-name=edge-tls

Remotely create a pair of certificate and private key with commonName edge:
	
	fabctl gen edge --api-server-address=http://host-cluster/
`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				util.Exitf("commonName is required")
			}

			if len(args) > 1 {
				util.Exitf("only one commonName is allowed")
			}

			for _, v := range certOptions.IPs {
				if net.ParseIP(v) == nil {
					util.Exitf("invalid IP: %s", v)
				}
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
			cli := newClient(clientGetter)

			var (
				caDER      []byte
				certDER    []byte
				keyDER     []byte
				err        error
				commonName = args[0]
			)

			if !commonOptions.Remote() {
				var caKeyDER []byte
				caDER, caKeyDER = cli.getCertAndKeyAsDER(commonOptions.CASecret)

				usages := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
				cfg := certOptions.AsConfig(commonName, false, usages)
				certDER, keyDER, err = certutil.NewCertFromCA2(caDER, caKeyDER, cfg)
				if err != nil {
					util.Exitf("failed to create certificate: %s\n", err)
				}
			} else {
				cacert, err := fclient.GetCertificate(commonOptions.APIServerAddress)
				if err != nil {
					util.Exitf("failed to get CA certificate from host cluster: %s\n", err)
				}

				var csrDER []byte
				keyDER, csrDER, err = certutil.NewCertRequest(certOptions.AsRequest(commonName))
				if err != nil {
					util.Exitf("failed to create certificate request: %s\n", err)
				}

				certPool := x509.NewCertPool()
				certPool.AddCert(cacert.Raw)
				cert, err := fclient.SignCertByToken(commonOptions.APIServerAddress, commonOptions.Token, csrDER, certPool)
				if err != nil {
					util.Exitf("failed to create certificate: %s\n", err)
				}
				caDER = cacert.DER
				certDER = cert.DER
			}

			secretName := commonName
			if len(saveOptions.SecretName) != 0 {
				secretName = saveOptions.SecretName
			}
			cli.saveCertAndKey(secretName, caDER, certDER, keyDER)
		},
	}

	fs := cmd.Flags()
	commonOptions.AddFlags(fs)
	certOptions.AddFlags(fs)
	saveOptions.AddFlags(fs)

	cmd.AddCommand(newCACmd(clientGetter))
	return cmd
}

func newCACmd(clientGetter types.ClientGetter) *cobra.Command {
	var certOptions CertOptions
	var secretName string

	cmd := &cobra.Command{
		Use:   "ca [CommonName]",
		Short: "Create a self-signed certificate and a private key",
		Long:  "Create a self-signed certificate and a private key, by default data will be save to a secret specified by '-ca-secret' flag.",
		Example: `Create a self-signed certificate and private key with default commonName:

	fabctl cert gen ca

Create a self-signed certificate and a private key with specified commonName:

	fabctl cert gen ca my-ca

Create a self-signed certificate and a private key, save data to secret ca-tls in namespace default:

	fabctl cert gen ca --ca-secret=ca-tls --namespace=default`,
		Args: cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			for _, v := range certOptions.IPs {
				if net.ParseIP(v) == nil {
					util.Exitf("Invalid IP: %s\n", v)
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cli := newClient(clientGetter)

			commonName := certutil.DefaultCAName
			if len(args) > 0 {
				commonName = args[0]
			}

			cfg := certOptions.AsConfig(commonName, true, nil)
			certDER, keyDER, err := certutil.NewSelfSignedCA(cfg)
			if err != nil {
				util.Exitf("failed to create certificate: %s\n", err)
			}

			cli.saveCAToSecret(secretName, certDER, keyDER)
		},
	}

	fs := cmd.Flags()
	certOptions.AddFlags(fs)
	fs.StringVar(&secretName, "secret-name", "fabedge-ca", "The name of the secret to store certificate and private key")

	return cmd
}
