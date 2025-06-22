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
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("cert.pem")
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(certData)

	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool, // use the cert pool with server's cert
			},
		},
	}
	resp, err := client.Get("https://localhost:8080/stream")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// read the response body
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		panic(err)
	}
}

func runServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// resposding with OK 200 header
		w.WriteHeader(http.StatusOK)
		// taking over the HTTP/3 stream
		streamer := w.(http3.HTTPStreamer)
		http3Stream := streamer.HTTPStream()
		defer http3Stream.Close()
		// sending data to the stream
		for i := range 10 {
			fmt.Fprintf(http3Stream, "data chunk #%d\n", i)
			time.Sleep(100 * time.Millisecond)
		}
	})
	srv := &http3.Server{
		// listen on the port 8080
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}
	// path to generated cert and key
	if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		panic(err)
	}
}
