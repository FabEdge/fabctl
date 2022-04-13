package cert

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/fabedge/fabctl/pkg/types"
	"github.com/fabedge/fabctl/pkg/util"
	"github.com/fabedge/fabedge/pkg/common/constants"
	certutil "github.com/fabedge/fabedge/pkg/util/cert"
	secretutil "github.com/fabedge/fabedge/pkg/util/secret"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type secretClient struct {
	*types.Client
}

func newClient(clientGetter types.ClientGetter) secretClient {
	cli, err := clientGetter.GetClient()
	util.CheckError(err)

	return secretClient{cli}
}

func (cli secretClient) saveCAToSecret(name string, caDER, keyDER []byte) {
	cli.createSecretIfNotExist(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cli.GetNamespace(),
			Labels: map[string]string{
				constants.KeyCreatedBy: "fabctl",
			},
		},
		Data: map[string][]byte{
			secretutil.KeyCACert: certutil.EncodeCertPEM(caDER),
			secretutil.KeyCAKey:  certutil.EncodePrivateKeyPEM(keyDER),
		},
	})
}

func (cli secretClient) saveCertAndKey(name string, caCertDER, certDER, keyDER []byte) {
	secret := secretutil.TLSSecret().
		Name(name).
		Namespace(cli.GetNamespace()).
		EncodeCACert(caCertDER).
		EncodeCert(certDER).
		EncodeKey(keyDER).
		Label(constants.KeyCreatedBy, "fabctl").
		Build()

	cli.createOrUpdateSecret(&secret)
}

func (cli secretClient) createOrUpdateSecret(secret *corev1.Secret) {
	err := cli.Create(context.TODO(), secret)
	switch {
	case err == nil:
		fmt.Printf("secret %s/%s is saved\n", secret.Namespace, secret.Name)
		return
	case errors.IsAlreadyExists(err):
		if err := cli.Update(context.TODO(), secret); err != nil {
			util.Exitf("failed to save secret: %s\n", err)
		}
		fmt.Printf("secret %s/%s is saved\n", secret.Namespace, secret.Name)
	default:
		util.Exitf("failed to save secret: %s\n", err)
	}
}

func (cli secretClient) createSecretIfNotExist(secret *corev1.Secret) {
	err := cli.Create(context.TODO(), secret)
	switch {
	case err == nil:
		fmt.Printf("secret %s/%s is saved\n", secret.Namespace, secret.Name)
		return
	case errors.IsAlreadyExists(err):
		fmt.Printf("secret %s/%s exists and gives up\n", secret.Namespace, secret.Name)
		return
	default:
		util.Exitf("failed to save secret: %s\n", err)
	}
}

func (cli secretClient) getCertAndKeyAsDER(secretName string) (certDER []byte, keyDER []byte) {
	secret := cli.getSecret(secretName)
	return cli.getCertAndKeyFromSecret(secret)
}

func (cli secretClient) getCertAndKeyFromSecret(secret corev1.Secret) (certDER []byte, keyDER []byte) {
	// CA TLS secret created by fabedge-cert CLI has ca.crt/ca.key fields, so
	// here we try to get data by keys  ca.crt and ca.key first
	certName, keyName := secretutil.KeyCACert, secretutil.KeyCAKey
	if secret.Data[certName] != nil && secret.Data[keyName] != nil {
		return decodePEM(secret.Data[certName]), decodePEM(secret.Data[keyName])
	}

	return decodePEM(secret.Data[corev1.TLSCertKey]), decodePEM(secret.Data[corev1.TLSPrivateKeyKey])
}

func (cli secretClient) getCertificate(secretName string) *x509.Certificate {
	secret := cli.getSecret(secretName)

	pemBytes := secret.Data[corev1.TLSCertKey]
	if len(pemBytes) == 0 {
		pemBytes = secret.Data[secretutil.KeyCACert]
	}

	cert, err := x509.ParseCertificate(decodePEM(pemBytes))
	if err != nil {
		util.Exitf("failed to decode certificate: %s\n", err)
	}

	return cert

}

func (cli secretClient) getSecret(name string) corev1.Secret {
	var (
		secret corev1.Secret
		key    = types.ObjectKey{Name: name, Namespace: cli.GetNamespace()}
	)

	err := cli.Get(context.TODO(), key, &secret)
	if err != nil {
		util.Exitf("failed to get secret: %s\n", err)
	}

	return secret
}

func decodePEM(data []byte) []byte {
	block, _ := pem.Decode(data)
	if block == nil {
		util.Exitf("failed to decode pem data\n")
	}

	return block.Bytes
}
