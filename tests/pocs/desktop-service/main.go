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
	"os/user"
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

func newStatus() (Status, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return Status{}, err
	}
	identity, err := user.Current()
	if err != nil {
		return Status{}, fmt.Errorf("resolve process identity: %w", err)
	}
	return Status{"poc.v0", version, os.Getpid(), hex.EncodeToString(id[:]), runtime.GOOS, identity.Username, time.Now().UTC().Format(time.RFC3339Nano)}, nil
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
	status, err := newStatus()
	if err != nil {
		return err
	}
	listener, err := listen(path)
	if err != nil {
		return err
	}
	defer listener.Close()
	var audit *os.File
	if auditPath := os.Getenv("HOPNODE_POC_AUDIT_FILE"); auditPath != "" {
		audit, err = os.OpenFile(auditPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return fmt.Errorf("create private POC audit: %w", err)
		}
		defer audit.Close()
	}
	server := newServer(status, audit)
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
	body, err := io.ReadAll(io.LimitReader(response.Body, 65537))
	if err != nil {
		return Status{}, err
	}
	if len(body) > 65536 {
		return Status{}, errors.New("Core response exceeded 64 KiB")
	}
	if err := json.Unmarshal(body, &status); err != nil {
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
