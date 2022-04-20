package cert

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/spf13/cobra"
)

func newViewCmd(clientGetter types.ClientGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view secretName",
		Short:   "View the certificate of a TLS secret",
		Example: "fabctl cert view secretName",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli := newClient(clientGetter)
			cert := cli.getCertificate(args[0])

			fmt.Printf("Version: %d\n", cert.Version)
			fmt.Printf("Subject: %s\n", cert.Subject)
			fmt.Printf("Issuer: %s\n", cert.Issuer)
			fmt.Printf("IsCA: %t\n", cert.IsCA)
			fmt.Printf("Signature Algorithm: %s\n", cert.SignatureAlgorithm)
			fmt.Printf("Publickey Algorithm: %s\n", cert.PublicKeyAlgorithm)
			fmt.Printf("Validity: \n")
			fmt.Printf("      Not Before: %s\n", cert.NotBefore)
			fmt.Printf("      Not After: %s\n", cert.NotAfter)
			fmt.Printf("Key length: %d\n", cert.PublicKey.(*rsa.PublicKey).Size()*8)
			fmt.Printf("Key Usage: %s\n", formatKeyUsage(cert.KeyUsage))
			fmt.Printf("Ext Key Usage: %s\n", formatExtUsages(cert.ExtKeyUsage))
			fmt.Printf("DNS Names: %s\n", strings.Join(cert.DNSNames, " "))
			fmt.Printf("IP Addresses: %s\n", formatIPs(cert.IPAddresses))
			fmt.Printf("Email Addresses: %s\n", strings.Join(cert.EmailAddresses, " "))
			fmt.Printf("URIs: %s\n", formatURIs(cert.URIs))
		},
	}

	return cmd
}

func formatKeyUsage(keyUsage x509.KeyUsage) string {
	var usages []string
	isUsed := func(expected x509.KeyUsage) bool {
		return keyUsage&expected == expected
	}

	if isUsed(x509.KeyUsageDigitalSignature) {
		usages = append(usages, "DigitalSignature")
	}

	if isUsed(x509.KeyUsageContentCommitment) {
		usages = append(usages, "ContentCommitment")
	}

	if isUsed(x509.KeyUsageKeyEncipherment) {
		usages = append(usages, "KeyEncipherment")
	}

	if isUsed(x509.KeyUsageDataEncipherment) {
		usages = append(usages, "DataEncipherment")
	}

	if isUsed(x509.KeyUsageKeyAgreement) {
		usages = append(usages, "KeyAgreement")
	}

	if isUsed(x509.KeyUsageCertSign) {
		usages = append(usages, "CertSign")
	}

	if isUsed(x509.KeyUsageCRLSign) {
		usages = append(usages, "CRLSign")
	}

	if isUsed(x509.KeyUsageEncipherOnly) {
		usages = append(usages, "EncipherOnly")
	}

	if isUsed(x509.KeyUsageDecipherOnly) {
		usages = append(usages, "DecipherOnly")
	}

	return strings.Join(usages, " ")
}

func formatExtUsages(extKeyUsages []x509.ExtKeyUsage) string {
	var usages []string

	for _, eku := range extKeyUsages {
		t := ""
		switch eku {
		case x509.ExtKeyUsageAny:
			t = "Any"
		case x509.ExtKeyUsageServerAuth:
			t = "ServerAuth"
		case x509.ExtKeyUsageClientAuth:
			t = "ClientAuth"
		case x509.ExtKeyUsageCodeSigning:
			t = "CodeSigning"
		case x509.ExtKeyUsageEmailProtection:
			t = "EmailProtection"
		case x509.ExtKeyUsageIPSECEndSystem:
			t = "IPSECEndSystem"
		case x509.ExtKeyUsageIPSECTunnel:
			t = "IPSECTunnel"
		case x509.ExtKeyUsageIPSECUser:
			t = "IPSECUser"
		case x509.ExtKeyUsageTimeStamping:
			t = "TimeStamping"
		case x509.ExtKeyUsageOCSPSigning:
			t = "OCSPSigning"
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			t = "MicrosoftServerGatedCrypto"
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			t = "NetscapeServerGatedCrypto"
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			t = "MicrosoftCommercialCodeSigning"
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			t = "MicrosoftKernelCodeSigning"
		}

		if t != "" {
			usages = append(usages, t)
		}
	}

	return strings.Join(usages, " ")
}

func formatIPs(ips []net.IP) string {
	var values []string
	for _, ip := range ips {
		values = append(values, ip.String())
	}

	return strings.Join(values, " ")
}

func formatURIs(urls []*url.URL) string {
	var values []string
	for _, url := range urls {
		values = append(values, url.String())
	}

	return strings.Join(values, " ")
}
