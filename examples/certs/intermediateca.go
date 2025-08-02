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
	rootPEMBytes, err := readPEMFile("root_ca_certificate.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read root CA certificate: %s\n", err)
		os.Exit(1)
	}
	rootCert, err := x509.ParseCertificate(rootPEMBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse root certificate: %s\n", err)
		os.Exit(1)
	}

	rootPrivateKeyBytes, err := readPEMFile("root_ca_private_key.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read root CA private key: %s\n", err)
		os.Exit(1)
	}

	rootPrivateKey, err := x509.ParsePKCS8PrivateKey(rootPrivateKeyBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse root CA private key: %s\n", err)
		os.Exit(1)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate serial number: %s\n", err)
		os.Exit(1)
	}

	intermediatePrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate private key: %s\n", err)
		os.Exit(1)
	}

	intermediateTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageCertSign,

		Subject: pkix.Name{
			Organization: []string{"Example Intermediate CA"},
		},

		// constraints
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &intermediateTemplate, rootCert, &intermediatePrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create certificate: %s\n", err)
		os.Exit(1)
	}

	if err := storeCertificate(certBytes, "intermediate_ca_certificate.pem"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to store certificate: %s\n", err)
		os.Exit(1)
	}
	if err := storePrivateKey(intermediatePrivateKey, "intermediate_ca_private_key.pem"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to store private key: %s\n", err)
		os.Exit(1)
	}
}

func readPEMFile(filename string) ([]byte, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read PEM file: %w", err)
	}

	// we know the file has only one PEM block
	pemBlock, _ := pem.Decode(fileBytes)
	if pemBlock == nil {
		return nil, fmt.Errorf("parse PEM: %w", err)
	}
	return pemBlock.Bytes, nil
}

func storeCertificate(cert []byte, filename string) error {
	certFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create certificate file: %w", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		return fmt.Errorf("write certificate to file: %w", err)
	}
	return nil
}

func storePrivateKey(key *ecdsa.PrivateKey, filename string) error {
	privateKeyFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}

	if err := pem.Encode(privateKeyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return fmt.Errorf("write private key to file: %w", err)
	}
	return nil
}
