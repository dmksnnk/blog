package main

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/dmksnnk/blog"
)

type config struct {
	ListenAddress string `env:"LISTEN_ADDRESS" envDefault:":8080"`
}

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := parseConfig()

	publicFS, err := fs.Sub(blog.Public, "public")
	if err != nil {
		slog.Error("failed to create sub fs", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/-/health", health())
	mux.Handle("/", http.FileServerFS(publicFS))

	srv := http.Server{
		Addr:    cfg.ListenAddress,
		Handler: mux,
	}

	go func() {
		slog.Info("starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	<-rootCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
	}

	slog.Info("server shutdown")
}

func parseConfig() config {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("parse config", "error", err)
		os.Exit(1)
	}

	return cfg
}

func health() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
