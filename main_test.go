package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

const drainGrace = 100 * time.Millisecond

func TestServeDrainsRequests(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "done")
	})

	ln, err := net.Listen(listenNetwork, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	served := make(chan error, 1)
	go func() { served <- serve(ctx, &http.Server{Handler: handler}, ln) }()

	replies := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			replies <- err.Error()
			return
		}
		defer func() { _ = resp.Body.Close() }()

		body, _ := io.ReadAll(resp.Body)
		replies <- string(body)
	}()

	<-started
	cancel()

	select {
	case err := <-served:
		t.Fatalf("serve returned with a request in flight: %v", err)
	case <-time.After(drainGrace):
	}

	close(release)

	if got := <-replies; got != "done" {
		t.Errorf("reply = %q, want the in-flight response", got)
	}
	if err := <-served; err != nil {
		t.Errorf("serve = %v, want nil", err)
	}
}
