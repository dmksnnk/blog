package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func main() {
	// path to generated cert and key
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping-pong", pingPongHandler)
	srv := &http3.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{http3.NextProtoH3}, // tell peer server supports HTTP/3 protocol
		},
		EnableDatagrams: true, // enable datagrams support
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("server error", "err", err)
				return
			}
		}
		slog.Info("server closed")
	}()

	if err := runClient(context.TODO()); err != nil {
		slog.Error("client error", "err", err)
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("server shutdown error", "err", err)
	}

	// wait for the server to exit
	wg.Wait()
}

func pingPongHandler(w http.ResponseWriter, r *http.Request) {
	conn := w.(http3.Hijacker).Connection()
	select {
	case <-conn.ReceivedSettings():
	case <-time.After(10 * time.Second):
		// didn't receive SETTINGS within 10 seconds
		http.Error(w, "timeout waiting for SETTINGS", http.StatusBadRequest)
		return
	}
	// check that HTTP Datagram support is enabled
	settings := conn.Settings()
	if !settings.EnableDatagrams {
		http.Error(w, "datagram support not enabled", http.StatusBadRequest)
		return
	}
	// resposding with OK 200 header
	w.WriteHeader(http.StatusOK)
	// taking over the HTTP/3 stream
	streamer := w.(http3.HTTPStreamer)
	http3Stream := streamer.HTTPStream()
	defer http3Stream.Close()

	ctx := r.Context()
	var (
		buf = make([]byte, 8)
		val uint64
		err error
	)
	// initialize ping-pong datagrams
	binary.BigEndian.PutUint64(buf, val)
	err = http3Stream.SendDatagram(buf)
	if err != nil {
		slog.Error("server: error sending initial datagram", "err", err)
		return
	}
	for {
		buf, err = http3Stream.ReceiveDatagram(ctx)
		if err != nil {
			slog.Error("server: error receiving datagram", "err", err)
			return
		}

		val = binary.BigEndian.Uint64(buf)
		slog.Info("server: received datagram", "value", val)

		// echoing back the datagram
		val++
		binary.BigEndian.PutUint64(buf, val)
		err = http3Stream.SendDatagram(buf)
		if err != nil {
			slog.Error("server: error sending datagram", "err", err)
			return
		}
		if val == 10 {
			slog.Info("server: sent final datagram, exiting")
			break
		}
	}
}

func runClient(ctx context.Context) error {
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile("cert.pem")
	if err != nil {
		return fmt.Errorf("read cert file: %w", err)
	}
	certPool.AppendCertsFromPEM(certData)
	tlsConf := &tls.Config{
		RootCAs:    certPool,                    // use the cert pool with server's cert
		NextProtos: []string{http3.NextProtoH3}, // use HTTP/3 protocol
	}
	quicConf := &quic.Config{
		EnableDatagrams: true, // enable QUIC datagrams support
	}
	tr := &http3.Transport{
		EnableDatagrams: true, // enable support for HTTP/3 datagrams
	}

	quicConn, err := quic.DialAddr(ctx, "localhost:8080", tlsConf, quicConf)
	if err != nil {
		return fmt.Errorf("dial QUIC: %w", err)
	}

	http3Conn := tr.NewClientConn(quicConn)
	// wait for the server's SETTINGS
	select {
	case <-http3Conn.ReceivedSettings():
	case <-http3Conn.Context().Done():
		// connection closed
		return fmt.Errorf("connection closed before receiving SETTINGS")
	}
	settings := http3Conn.Settings()
	if !settings.EnableDatagrams {
		// no datagram support, closing connection
		http3Conn.CloseWithError(http3.ErrCodeNoError, "datagram support not enabled")
		return fmt.Errorf("server does not support datagrams")
	}
	defer http3Conn.CloseWithError(http3.ErrCodeNoError, "bye!")

	reqStream, err := http3Conn.OpenRequestStream(ctx)
	if err != nil {
		return fmt.Errorf("open request stream: %w", err)
	}
	defer reqStream.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://localhost:8080/ping-pong", http.NoBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if err := reqStream.SendRequestHeader(req); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	resp, err := reqStream.ReadResponse()
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var (
		val uint64
		buf = make([]byte, 8)
	)
	for {
		buf, err = reqStream.ReceiveDatagram(ctx)
		if err != nil {
			return fmt.Errorf("receive datagram: %w", err)
		}

		val = binary.BigEndian.Uint64(buf)
		slog.Info("client: received datagram", "value", val)
		if val == 10 {
			slog.Info("client: received final datagram, exiting")
			return nil
		}

		val++
		binary.BigEndian.PutUint64(buf, val)
		err = reqStream.SendDatagram(buf)
		if err != nil {
			return fmt.Errorf("send datagram: %w", err)
		}
	}
}
