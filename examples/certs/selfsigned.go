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
	"net"
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

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate private key: %s\n", err)
		os.Exit(1)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		// SANs
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost"},
	}

	// Parent is equal to template, which means it is self-signed.
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create certificate: %s\n", err)
		os.Exit(1)
	}

	privateKeyFile, err := os.Create("private_key.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create private key file: %s\n", err)
		os.Exit(1)
	}
	defer privateKeyFile.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal private key: %s\n", err)
		os.Exit(1)
	}

	if err := pem.Encode(privateKeyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write private key to file: %s\n", err)
		os.Exit(1)
	}

	certFile, err := os.Create("self_signed_certificate.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create certificate file: %s\n", err)
		os.Exit(1)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write certificate to file: %s\n", err)
		os.Exit(1)
	}
}
