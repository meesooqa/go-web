package web

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestServer_Run_StartsServesAndShutsDownGracefully(t *testing.T) {
	c := testConfig()
	c.ShutdownTimeout = 2 * time.Second

	srv, err := New(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	srv.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	ctx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() { runErr <- srv.Run(ctx) }()

	select {
	case <-srv.ready:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start listening within the allotted time")
	}

	if srv.Addr() == "" {
		t.Fatal("Addr() is empty after server is ready")
	}

	resp, err := http.Get("http://" + srv.Addr() + "/ping")
	if err != nil {
		t.Fatalf("error requesting the running server: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, expected 200", resp.StatusCode)
	}

	cancel()

	select {
	case err := <-runErr:
		if err != nil {
			t.Errorf("Run returned an error after context cancellation: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not exit after context cancellation, graceful shutdown hung")
	}
}

func TestServer_Run_InvalidAddress(t *testing.T) {
	c := testConfig()
	c.Host = "not-a-valid-host-!@#"
	c.Port = 99999 // out of valid port range

	srv, err := New(c, nil)
	if err != nil {
		t.Fatalf("unexpected error in New: %v", err)
	}

	err = srv.Run(context.Background())
	if err == nil {
		t.Fatal("expected an error for an invalid listen address")
	}
}
