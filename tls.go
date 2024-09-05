package x

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"gitlab.com/tozd/go/errors"
)

// CreateTempCertificateFiles creates a pair of files for given domains.
// It generates a ECDSA private key.
func CreateTempCertificateFiles(certPath, keyPath string, domains []string) errors.E {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return errors.WithStack(err)
	}

	// Create a self-signed certificate.
	template := x509.Certificate{ //nolint:exhaustruct
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Test"}}, //nolint:exhaustruct
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(24 * time.Hour), //nolint:mnd
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              domains,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return errors.WithStack(err)
	}

	// Write the certificate to a file.
	certFile, err := os.Create(certPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}) //nolint:exhaustruct
	if err != nil {
		return errors.WithStack(err)
	}

	// Write the private key to a file.
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer keyFile.Close()
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return errors.WithStack(err)
	}

	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}) //nolint:exhaustruct
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
