package cert

import (
	"crypto/x509"
	"net"

	certutil "github.com/fabedge/fabedge/pkg/util/cert"
	timeutil "github.com/fabedge/fabedge/pkg/util/time"
	flag "github.com/spf13/pflag"
)

type CommonOptions struct {
	CASecret string

	APIServerAddress string
	Token            string
}

func (opts *CommonOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&opts.CASecret, "ca-secret", "fabedge-ca", "The name of ca secret, by default CLI read CA cert/key from secret")

	fs.StringVar(&opts.APIServerAddress, "api-server-address", "", "The address of host cluster's API server, when this option is set, generate or verify certificate remotely")
	fs.StringVar(&opts.Token, "token", "", "Authentication token, not necessary when verifying certificate")
}

func (opts *CommonOptions) Remote() bool {
	return len(opts.APIServerAddress) != 0
}

type CertOptions struct {
	Organization   []string
	ValidityPeriod int64
	IPs            []string
	DNSNames       []string
}

func (opts *CertOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringSliceVarP(&opts.Organization, "organization", "O", []string{certutil.DefaultOrganization}, "your organization name")
	fs.Int64Var(&opts.ValidityPeriod, "validity-period", 365, "validity period for your cert, unit: day")
	fs.StringSliceVar(&opts.IPs, "ips", nil, "The ip addresses for your cert, e.g. 2.2.2.2,10.10.10.10")
	fs.StringSliceVar(&opts.DNSNames, "dns-names", nil, "The dns names for your cert, e.g. fabedge.io,yourdomain.com")
}

func (opts *CertOptions) AsConfig(cn string, isCA bool, usages []x509.ExtKeyUsage) certutil.Config {
	return certutil.Config{
		CommonName:     cn,
		IsCA:           isCA,
		Organization:   opts.Organization,
		IPs:            opts.GetIPs(),
		DNSNames:       opts.DNSNames,
		ValidityPeriod: timeutil.Days(opts.ValidityPeriod),
		Usages:         usages,
	}
}

func (opts *CertOptions) AsRequest(cn string) certutil.Request {
	return certutil.Request{
		CommonName:   cn,
		Organization: opts.Organization,
		IPs:          opts.GetIPs(),
		DNSNames:     opts.DNSNames,
	}
}

func (opts *CertOptions) GetIPs() (ips []net.IP) {
	for _, v := range opts.IPs {
		ips = append(ips, net.ParseIP(v))
	}

	return ips
}

type SaveOptions struct {
	SecretName string
}

func (opts *SaveOptions) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&opts.SecretName, "secret-name", "", "The name of the secret to store certificate and private key, if not provided, the commonName will be used")
}
