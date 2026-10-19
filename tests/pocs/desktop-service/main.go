// Purpose: independent controller and headless CLI IPC fixture. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

const version = "0.1.0-poc"

type Status struct {
	APIVersion string `json:"api_version"`
	Version    string `json:"version"`
	PID        int    `json:"pid"`
	InstanceID string `json:"instance_id"`
	OS         string `json:"os"`
	Identity   string `json:"identity"`
	StartedAt  string `json:"started_at"`
}

func newStatus() Status {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	return Status{"poc.v0", version, os.Getpid(), hex.EncodeToString(id[:]), runtime.GOOS, "user", time.Now().UTC().Format(time.RFC3339Nano)}
}

func handler(status Status) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v0/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(status)
	})
	mux.HandleFunc("GET /v0/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"api_version": "poc.v0", "version": version})
	})
	return mux
}

func serve(ctx context.Context, path string) error {
	listener, err := listen(path)
	if err != nil {
		return err
	}
	defer listener.Close()
	server := &http.Server{Handler: handler(newStatus()), ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 3 * time.Second, MaxHeaderBytes: 4096}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	fmt.Fprintln(os.Stderr, "HopNode P04 controller ready (private local IPC)")
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func query(ctx context.Context, path string) (Status, error) {
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) { return dial(ctx, path) }}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/v0/status", nil)
	if err != nil {
		return Status{}, err
	}
	response, err := client.Do(req)
	if err != nil {
		return Status{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Status{}, fmt.Errorf("Core returned HTTP %d", response.StatusCode)
	}
	var status Status
	if err := json.NewDecoder(io.LimitReader(response.Body, 65537)).Decode(&status); err != nil {
		return Status{}, err
	}
	if status.APIVersion != "poc.v0" || status.PID <= 0 || status.InstanceID == "" {
		return Status{}, errors.New("incompatible or incomplete Core status")
	}
	return status, nil
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: hopnodectl-poc serve|status|doctor|--version [--json]")
	}
	if args[0] == "--version" {
		fmt.Println("hopnodectl-poc " + version)
		return nil
	}
	path, err := endpoint()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	switch args[0] {
	case "serve":
		return serve(ctx, path)
	case "status", "doctor":
		status, err := query(ctx, path)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(status)
	default:
		return fmt.Errorf("unknown POC command %q", args[0])
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
