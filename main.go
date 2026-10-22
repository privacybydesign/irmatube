package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	listenNetwork     = "tcp"
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

	ln, err := net.Listen(listenNetwork, listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("listening on %s", listenAddr)
	if err := serve(ctx, srv, ln); err != nil {
		log.Fatal(err)
	}
}

// serve runs srv on ln until ctx is done, then waits for in-flight requests.
func serve(ctx context.Context, srv *http.Server, ln net.Listener) error {
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Serve returns as soon as Shutdown starts; wait for the drain.
	<-shutdownDone

	return nil
}
