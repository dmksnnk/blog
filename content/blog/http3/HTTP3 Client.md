---
date: '2025-05-08T19:04:26+02:00'
draft: true
title: 'HTTP3 Client'
slug: 'http3-client'
showToc: true
tags:
  - HTTP3
---

> Previous parts:
> - [Writing HTTP3 Server](/blog/http3/writing-http3-server/)

Today we will create HTTP/3 client to interact with our [HTTP/3 Server](/blog/http3/writing-http3-server/).

First of all you'll need the running HTTP/3 server. Run it in separate terminal:

```sh
go run server.go
```

## Client skipping TLS verification

Let's write a simple HTTP/3 client for our server. It is pretty straightforward: we need to create a new HTTP/3 transport and pass the transport to HTTP client from `net/http`:

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

The rest is the same as in regular `net/http` client. Let's test it by calling our server:

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

Running it all together should yield next:

```sh
$ go run client.go
Response status: 200 OK
Response body: Hello, World!
```


## Trusting Server's TLS Certificate

Load server's certificate generated [[Writing HTTP3 Server#Generating Certificate|in the previous step]].

```go filename=client.go
// load server's cert in PEM format
certPem, err := os.ReadFile("cert.pem")
if err != nil {
    panic(err)
}
```

As it is PEM-endoded data, we need to decode it and take only CERTIFICATE block from it.

```go filename=client.go
// there is only one block CERTIFICATE in the cert file
certRaw, _ := pem.Decode(certPem) // decode the PEM encoded cert
cert, err := x509.ParseCertificate(certRaw.Bytes)
if err != nil {
    panic(err)
}
```

Now, create a new certificate pool with loaded certificate for our client to use:

```go filename=client.go
// create a cert pool and add the server's cert
certPool := x509.NewCertPool()
certPool.AddCert(cert)
```

Pass the certificate pool to the client's TLS config. Note, there is no `InsecureSkipVerify` option, so our client will use certificates from the pool to validate server's certificate. T

```go filename=client.go
tr := &http3.Transport{
    TLSClientConfig: &tls.Config{
        RootCAs: certPool, // use the cert pool with server's cert
    },
}
```

The rest is the same, pass the transport to the client and try to call the server:

```go filename=client.go
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
```

You should receive the same output as before.

## Recap

Today we have learned how to write a simple HTTP/3 client, verified that it works.
Then, we have leaned how to configure HTTP/3 client to trust the server's certificate.

You can find additional information in the official [quic-go](https://quic-go.net/docs/http3/client/) documentation.

> **Next Parts:**
>
> - Stream data from HTTP/3 server.
> - Stream data from an HTTP/3 client.
> - Send HTTP/3 DATAGRAMs.