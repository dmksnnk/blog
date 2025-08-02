package main

import (
	"crypto"
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
	intermediatePEMBytes, err := readPEMFile("intermediate_ca_certificate.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read intermediate CA certificate: %s\n", err)
		os.Exit(1)
	}
	intermediateCert, err := x509.ParseCertificate(intermediatePEMBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse intermediate certificate: %s\n", err)
		os.Exit(1)
	}

	intermediatePrivateKeyBytes, err := readPEMFile("intermediate_ca_private_key.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read intermediate CA private key: %s\n", err)
		os.Exit(1)
	}

	intermediatePrivateKey, err := x509.ParsePKCS8PrivateKey(intermediatePrivateKeyBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse intermediate CA private key: %s\n", err)
		os.Exit(1)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate private key: %s\n", err)
		os.Exit(1)
	}

	csr := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "example.com",
		},
		DNSNames: []string{"example.com", "www.example.com"},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csr, privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create certificate request: %s\n", err)
		os.Exit(1)
	}

	certBytes, err := createCert(csrBytes, intermediateCert, intermediatePrivateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create certificate: %s\n", err)
		os.Exit(1)
	}

	if err := storeCertificate(certBytes, "certificate.pem"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write certificate: %s\n", err)
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

func createCert(csrBytes []byte, intermediateCert *x509.Certificate, intermediatePrivateKey crypto.PrivateKey) ([]byte, error) {
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate request: %w", err)
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("check certificate request signature: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		Subject:      csr.Subject,

		// SANs
		DNSNames:       csr.DNSNames,
		EmailAddresses: csr.EmailAddresses,
		IPAddresses:    csr.IPAddresses,
		URIs:           csr.URIs,
	}

	return x509.CreateCertificate(rand.Reader, template, intermediateCert, csr.PublicKey, intermediatePrivateKey)
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
