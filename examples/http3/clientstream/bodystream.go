package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func main() {
	go runServer()
	time.Sleep(100 * time.Millisecond)

	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("cert.pem")
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(certData)

	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    certPool, // use the cert pool with server's cert
				NextProtos: []string{http3.NextProtoH3},
			},
		},
	}

	pipeR, pipeW := io.Pipe()
	req, err := http.NewRequest("POST", "https://localhost:8080/stream", pipeR)
	if err != nil {
		panic(err)
	}

	go func() {
		defer req.Body.Close()

		for i := 0; i < 10; i++ {
			if _, err := fmt.Fprintf(pipeW, "data chunk #%d\n", i); err != nil {
				panic(err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}
}

func runServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(os.Stdout, r.Body)
	})

	srv := &http3.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	err := srv.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}
}
