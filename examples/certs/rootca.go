package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func main() {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate serial number: %s\n", err)
		os.Exit(1)
	}

	rootPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate private key: %s\n", err)
		os.Exit(1)
	}

	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageCertSign,
		Subject: pkix.Name{
			Organization: []string{"Example Corp"},
		},
		// constraints
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create certificate: %s\n", err)
		os.Exit(1)
	}

	if err := storeCertificate(certBytes, "root_ca_certificate.pem"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to store certificate: %s\n", err)
		os.Exit(1)
	}
	if err := storePrivateKey(rootPrivateKey, "root_ca_private_key.pem"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to store private key: %s\n", err)
		os.Exit(1)
	}
}

func storeCertificate(cert []byte, filename string) error {
	certFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		return fmt.Errorf("failed to write certificate to file: %w", err)
	}
	return nil
}

func storePrivateKey(key *ecdsa.PrivateKey, filename string) error {
	privateKeyFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(privateKeyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key to file: %w", err)
	}
	return nil
}
