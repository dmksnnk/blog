package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/quic-go/quic-go"
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

	tlsConf := &tls.Config{
		RootCAs:    certPool,
		NextProtos: []string{http3.NextProtoH3},
	}

	qconn, err := quic.DialAddr(context.TODO(), "localhost:8080", tlsConf, nil)
	if err != nil {
		panic(err)
	}

	tr := &http3.Transport{}
	clientConn := tr.NewClientConn(qconn)
	stream, err := clientConn.OpenRequestStream(context.TODO())
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	req, err := http.NewRequest("POST", "https://localhost:8080/stream", http.NoBody)
	if err != nil {
		panic(err)
	}
	// sending request headers
	if err := stream.SendRequestHeader(req); err != nil {
		panic(err)
	}
	resp, err := stream.ReadResponse()
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	for i := 0; i < 10; i++ {
		fmt.Fprintf(stream, "data chunk #%d\n", i)
		time.Sleep(100 * time.Millisecond)
	}
}

func runServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// responding with 200 OK header
		w.WriteHeader(http.StatusOK)
		// taking over the HTTP/3 stream
		streamer := w.(http3.HTTPStreamer)
		http3Stream := streamer.HTTPStream()
		defer http3Stream.Close()
		// dumping the stream to stdout
		io.Copy(os.Stdout, http3Stream)
	})
	srv := &http3.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}
	// path to generated cert and key
	if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		panic(err)
	}
}
