package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	listenAddr        = ":8080"
	readHeaderTimeout = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

//go:embed static
var staticFiles embed.FS

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	videos, err := loadVideos(videoFiles)
	if err != nil {
		log.Fatal(err)
	}

	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           newServer(newIrmaClient(cfg), newVerifier(cfg, videos), static),
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("listening on %s", listenAddr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
