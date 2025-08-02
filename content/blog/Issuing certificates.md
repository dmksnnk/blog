---
date: '2025-08-01T13:30:00+02:00'
title: 'Issuing TLS Certificates in Go'
slug: 'tls-certificates'
showToc: true
draft: false
cover:
    image: 'certificates.svg'
summary: |
    A practical guide to issuing TLS certificates in Go.
    Learn how to create self-signed certificates,
    set up a Certificate Authority (CA), establish a trust chain,
    and issue certificates from a Certificate Signing Request (CSR).'
tags:
    - Go
    - x509
    - TLS
---

I don't have in-depth knowledge of how TLS works under the hood.
This post is a practical guide with experiments to learn more about creating certificates.

We will cover how to create a self-signed certificate,
a Certificate Authority (CA) to issue certificates, how to create a certificate trust chain
and how to issue certificates from a Certificate Signing Request (CSR).

Certificates contain various information, such as a public key, identification information for
the entity to whom the certificate was issued, and a digital signature.
When a client receives a certificate, it can validate that the certificate is signed by a trusted authority
and is issued to the expected entity. In the case of HTTPS, when a client (e.g., a  web browser) receives a server's certificate,
it verifies that the certificate is signed by a trusted CA and that the DNS name in it matches the requested host.

# Self-signed certificates

Self-signed certificates are certificates that are signed by the entity itself.
This is mainly useful for testing. First, we will generate a random serial number:

```go
serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
```

Next, we'll create a new private key:

```go
privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
```

This will generate an Elliptic Curve Digital Signature Algorithm (ECDSA) private key.
Another option is to use [Ed25519](https://pkg.go.dev/crypto/ed25519).

Now, let's create a certificate template:

```go
template := x509.Certificate{
    SerialNumber: serialNumber,
    NotBefore:    time.Now(),
    NotAfter:     time.Now().Add(time.Hour),
    KeyUsage:     x509.KeyUsageDigitalSignature,
    ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
    Subject: pkix.Name{
		CommonName:   "localhost",
    },
    // SANs
    IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
    DNSNames:    []string{"localhost"},
}
```
Let's break down what we are creating here:
- `SerialNumber` - A unique number for the certificate.
- `NotBefore` - The certificate is not valid before this time. In our case, we are making it valid from now.
- `NotAfter` - The certificate is not valid after this time. We are making it valid for one hour.
- `KeyUsage` - A restriction on how this certificate can be used. For a web server, we will need `digitalSignature`.
    See [RFC 5280 section 4.2.1.3.](https://www.rfc-editor.org/rfc/rfc5280#section-4.2.1.3) for more details.
- `ExtKeyUsage` - An additional application-specific purpose. In our case, it is for web server authentication.
- `Subject` - The Distinguished Name (DN) of the entity. For websites, this is usually the
domain name in the Common Name (CN).

The next fields are Subject Alternative Names (SANs), which define the identity of the certificate holder.
This can be DNS names (most commonly used for domain names), IP addresses, email addresses, or URIs.
For local testing, we are creating a certificate for the `localhost` domain and the loopback IPv4 and IPv6 addresses.

The next part is to create the certificate itself. It is created from the template, the parent certificate,
the public key of the certificate requestor, and the private key of the certificate signer.
As we are creating a self-signed certificate, we are passing the template as the parent and our own private key to sign it.

```go
certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
```

That's it, we have a private key and a certificate. Let's store them and inspect what `openssl`
can tell us about it. Certificates and keys are commonly stored as
[PEM-encoded](https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail) data,
which is `base64`-encoded binary data with labels indicating the beginning and the end.
Private keys are stored with the label `PRIVATE KEY` and certificates with `CERTIFICATE`.

```go
privateKeyFile, err := os.Create("private_key.pem")
...
privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
...
err = pem.Encode(privateKeyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})
...

certFile, err := os.Create("self_signed_certificate.pem")
...
err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
...
```

If you run this code with `go run ./selfsigned.go`, it will generate and store the
certificate and private key.
See the [full example on GitHub](https://github.com/dmksnnk/blog/tree/main/examples/certs/selfsigned.go).

To inspect the certificate, we can use `openssl`:

```sh
openssl x509 -in self_signed_certificate.pem -text -noout
```

{{< details summary="openssl output" >}}

```
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            87:3a:90:03:57:18:83:0d:ed:52:4c:ae:de:d2:11:a6
        Signature Algorithm: ecdsa-with-SHA256
        Issuer:
        Validity
            Not Before: Jul 31 20:41:35 2025 GMT
            Not After : Jul 31 21:41:35 2025 GMT
        Subject:
        Subject Public Key Info:
            Public Key Algorithm: id-ecPublicKey
                Public-Key: (256 bit)
                pub:
                    04:d8:a0:a5:6f:be:0d:69:b2:65:1f:cc:53:dc:13:
                    ad:b9:34:cb:bd:f2:47:d9:8a:0c:c0:09:94:c5:0b:
                    f2:d6:a2:21:9d:93:20:9f:61:85:ca:8b:da:ee:20:
                    ec:a2:af:b6:6f:af:b3:52:d1:33:75:55:20:f9:67:
                    c7:af:14:b4:e2
                ASN1 OID: prime256v1
                NIST CURVE: P-256
        X509v3 extensions:
            X509v3 Key Usage: critical
                Digital Signature
            X509v3 Extended Key Usage:
                TLS Web Server Authentication
            X509v3 Subject Alternative Name: critical
                DNS:localhost, IP Address:127.0.0.1, IP Address:0:0:0:0:0:0:0:1
    Signature Algorithm: ecdsa-with-SHA256
    Signature Value:
        30:45:02:20:70:3e:ab:7d:ea:71:d0:10:55:dd:61:f8:91:be:
        3f:e4:37:06:d2:6f:de:b2:7f:ad:a3:51:f0:68:b0:19:c7:73:
        02:21:00:b8:1f:bc:2f:68:aa:f6:f8:32:6f:d9:89:6c:2f:56:
        60:cd:a4:3b:45:d9:ea:1f:9b:2d:a4:49:70:7d:41:86:69
```

{{< /details >}}


# Root CA Certificate

Now that we have covered the basics, let's create a Root CA certificate.
This will be a self-signed certificate, much like the previous one, but with some key differences:

```go
template := x509.Certificate{
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
```

See the [full example on GitHub](https://github.com/dmksnnk/blog/tree/main/examples/certs/rootca.go).

The key differences are:
 - `KeyUsage` is set to `KeyUsageCertSign`, which means the public key in this certificate can be used to
    verify signatures on other certificates.
 - We are also specifying that this is a CA certificate by setting `IsCA` to `true`.
    `BasicConstraintsValid` is a flag that indicates that the constraints are valid and not just a zero value.

The private key of the CA certificate will be used to sign new intermediate CA certificates
or end-entity (e.g., web server) certificates.

If you inspect the generated certificate:

```sh
openssl x509 -in intermediate_ca_certificate.pem -text -noout
```

You can see that it is a CA certificate:

```
X509v3 extensions:
    X509v3 Key Usage: critical
        Certificate Sign
    X509v3 Basic Constraints: critical
        CA:TRUE
```

# Intermediate CA Certificate

Now that we have our new root CA, we can create intermediate CAs.
This means our root certificate can issue new CA certificates for intermediate CAs,
which in turn can issue certificates for end-entities, creating a chain of trust.

```
┌───────┐       ┌───────────────┐       ┌───────────────┐
│Root CA│─Signs─▶Intermediate CA├─Signs─▶End-entity cert│
└───────┘       └───────────────┘       └───────────────┘
```

Using the root CA's certificate and private key, we can create and sign a
new intermediate certificate. In the template, we are specifying
the usage and that this is a CA certificate.

```go
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
```

And creating the certificate itself:

```go
certBytes, err := x509.CreateCertificate(rand.Reader, &intermediateTemplate, rootCert, &intermediatePrivateKey.PublicKey, rootPrivateKey)
```

See the [full example on GitHub](https://github.com/dmksnnk/blog/tree/main/examples/certs/intermediateca.go).

Note that we are using the root certificate as the parent certificate
and signing with the root's private key, which indicates
that the root CA has verified the identity of the public key holder and trusts it.

If we check the certificate:

```sh
openssl x509 -in intermediate_ca_certificate.pem -text -noout
```

You can see that the issuer is our root CA's `Example Corp`:

```
Issuer: O=Example Corp
```

We can also verify the whole chain of certificates:

```sh
openssl verify -show_chain -verbose -x509_strict -CAfile root_ca_certificate.pem  intermediate_ca_certificate.pem
```

`openssl` tells us that the intermediate certificate is OK and shows us the certificate chain:

```
intermediate_ca_certificate.pem: OK
Chain:
depth=0: O=Example Intermediate CA (untrusted)
depth=1: O=Example Corp
```

# Certificate Signing Request

A [Certificate Signing Request](https://en.wikipedia.org/wiki/Certificate_signing_request) (CSR)
is a message you send to a CA to obtain a certificate.
It contains a public key, identifying information (e.g., domain name, IP address), and
is signed by your private key. By having the public key in the CSR, the CA can verify that the request was
not tampered with. Next, the CA will also verify the claims in the CSR. For example,
it will verify that you are the owner of the domain by asking you to perform "challenges".
Here is a list of [Let's Encrypt challenges](https://letsencrypt.org/docs/challenge-types/).

Let's create a new CSR. First, we need a private key:

```go
privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
...
```

And a request itself:

```go
csr := &x509.CertificateRequest{
    Subject: pkix.Name{
        CommonName: "example.com",
    },
    DNSNames: []string{"example.com", "www.example.com"},
}
```

We are asking for a certificate for the `example.com` domain, and to cover common use-cases,
we are asking for `example.com` and `www.example.com` domains.

Now, sign it with our private key:

```go
csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csr, privateKey)
...
```

Now that we have our CSR, we need to pass it to a CA, where it can verify our claims and issue a new
certificate. On the CA side, we would parse the CSR:

```go
csr, err := x509.ParseCertificateRequest(csrBytes)
...
```
And ensure the request was not tampered with by checking its signature:

```go
err := csr.CheckSignature()
...
```

Here the CA would also verify that you are who you claim to be in the CSR
by performing challenges mentioned before. We are skipping this part.

Prepare a new template. At this stage, we can also add additional information to the template,
such as user info.

```go
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
```

Create the certificate itself. Note that we are using our intermediate CA and its private key to sign
the certificate.

```go
cert, err := x509.CreateCertificate(rand.Reader, template, intermediateCert, csr.PublicKey, intermediatePrivateKey)
...
```

See the [full example on GitHub](https://github.com/dmksnnk/blog/tree/main/examples/certs/csr.go).

Now, let's check what `openssl` will tell us about the whole chain,
from the root certificate to the newly created end-user certificate:

```sh
$ openssl verify -show_chain -verbose -x509_strict -CAfile root_ca_certificate.pem -untrusted intermediate_ca_certificate.pem certificate.pem
certificate.pem: OK
Chain:
depth=0: CN=example.com (untrusted)
depth=1: O=Example Intermediate CA (untrusted)
depth=2: O=Example Corp
```

And that's it. We have covered the whole path from issuing a self-signed certificate
to creating a chain of trust to issue certificates from CSRs.
