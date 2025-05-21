---
date: '2025-05-08T19:04:26+02:00'
title: 'HTTP3 Client'
slug: 'http3-client'
showToc: true
tags:
  - HTTP3
---

> Previous parts:
> - [Writing HTTP3 Server](/blog/http3/http3-server/)

Today, we will create an HTTP/3 client to interact with our [HTTP/3 Server](/blog/http3/http3-server/).

First of all, you'll need a running HTTP/3 server. Run it in a separate terminal:

```sh
go run server.go
```

## Client skipping TLS verification

Let's write a simple HTTP/3 client for our server. It's pretty straightforward: we need to create a new HTTP/3 transport and pass it to an HTTP client from `net/http`:

```go filename=client.go
// client.go
tr := &http3.Transport{
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true, // skip TLS verification
    },
}
client := &http.Client{
    Transport: tr,
}
```

The rest is the same as with a regular `net/http` client. Let's test it by calling our server:

```go filename=client.go
// client.go
resp, err := client.Get("https://localhost:8080")
if err != nil {
    panic(err)
}
defer resp.Body.Close()

fmt.Printf("Response status: %s\n", resp.Status)
body, err := io.ReadAll(resp.Body) // read the response body
if err != nil {
    panic(err)
}

fmt.Printf("Response body: %s\n", body) // print the response body
```

{{< details summary="Full code" >}}

```go
// client.go
package main

import (
    "crypto/tls"
    "fmt"
    "io"
    "net/http"

    "github.com/quic-go/quic-go/http3"
)

func main() {
    tr := &http3.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // skip TLS verification
        },
    }
    client := &http.Client{
        Transport: tr,
    }
    resp, err := client.Get("https://localhost:8080")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Printf("Response status: %s\n", resp.Status)
    body, err := io.ReadAll(resp.Body) // read the response body
    if err != nil {
        panic(err)
    }

    fmt.Printf("Response body: %s\n", body) // print the response body
}

```
{{< /details >}}


Running it all together should yield the following:

```sh
$ go run client.go
Response status: 200 OK
Response body: Hello, World!
```

## Trusting Server's TLS Certificate

Load server's certificate generated [in the previous step](/blog/http3/http3-server/#generating-certificate).

```go filename=client.go
// load server's cert in PEM format
certPem, err := os.ReadFile("cert.pem")
if err != nil {
    panic(err)
}
```

As it is PEM-encoded data, we need to decode it and extract only the CERTIFICATE block.

```go
// there is only one block CERTIFICATE in the cert file
certRaw, _ := pem.Decode(certPem) // decode the PEM encoded cert
cert, err := x509.ParseCertificate(certRaw.Bytes)
if err != nil {
    panic(err)
}
```

Now, create a new certificate pool with the loaded certificate for our client to use:

```go
// create a cert pool and add the server's cert
certPool := x509.NewCertPool()
certPool.AddCert(cert)
```

Pass the certificate pool to the client's TLS config. Note that there is no `InsecureSkipVerify` option, so our client will use certificates from the pool to validate the server's certificate.

```go filename=client.go
tr := &http3.Transport{
    TLSClientConfig: &tls.Config{
        RootCAs: certPool, // use the cert pool with server's cert
    },
}
```

The rest is the same, pass the transport to the client and try to call the server.
And if you run it againg, you should receive the same output as before. You are amazing 🎉!

{{< details summary="Full code" >}}

```go
// client.go
package main

import (
    "crypto/tls"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io"
    "net/http"
    "os"

    "github.com/quic-go/quic-go/http3"
)

func main() {
    // load server's cert in PEM format
    certPem, err := os.ReadFile("cert.pem")
    if err != nil {
        panic(err)
    }

    // there is only one block CERTIFICATE in the cert file
    certRaw, _ := pem.Decode(certPem) // decode the PEM encoded cert
    cert, err := x509.ParseCertificate(certRaw.Bytes)
    if err != nil {
        panic(err)
    }
    // create a cert pool and add the server's cert
    certPool := x509.NewCertPool()
    certPool.AddCert(cert)

    tr := &http3.Transport{
        TLSClientConfig: &tls.Config{
            RootCAs: certPool, // use the cert pool with server's cert
        },
    }
    client := &http.Client{
        Transport: tr,
    }
    resp, err := client.Get("https://localhost:8080")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Printf("Response status: %s\n", resp.Status)
    body, err := io.ReadAll(resp.Body) // read the response body
    if err != nil {
        panic(err)
    }

    fmt.Printf("Response body: %s\n", body) // print the response body
}

```
{{< /details >}}



## Recap

Today we have learned how to write a simple HTTP/3 client, verified that it works.
Then, we have leaned how to configure HTTP/3 client to trust the server's certificate.

You can find additional information in the official [quic-go](https://quic-go.net/docs/http3/client/) documentation.

> **Next Parts:**
>
> - Stream data from HTTP/3 server.
> - Stream data from an HTTP/3 client.
> - Send HTTP/3 DATAGRAMs.