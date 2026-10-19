// Purpose: real local transport roundtrip and HTTP surface tests. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRealTransportRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "controller.sock")
	if runtime.GOOS == "windows" {
		path = fmt.Sprintf(`\\.\pipe\hopnode-poc-test-%d-%d`, os.Getpid(), time.Now().UnixNano())
	} else {
		if err := os.Chmod(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
	}
	listener, err := listen(path)
	if err != nil {
		t.Fatal(err)
	}
	status, err := newStatus()
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: handler(status), ReadHeaderTimeout: time.Second}
	t.Cleanup(func() { _ = server.Close() })
	go func() { _ = server.Serve(listener) }()
	got, err := query(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got != status {
		t.Fatalf("received %+v; expected %+v", got, status)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Fatalf("socket mode %v", info.Mode())
		}
	}
}

func TestReadOnlyHTTP(t *testing.T) {
	status, err := newStatus()
	if err != nil {
		t.Fatal(err)
	}
	h := handler(status)
	for _, test := range []struct {
		method, path string
		code         int
	}{{"GET", "/v0/status", 200}, {"GET", "/v0/version", 200}, {"POST", "/v0/status", 405}, {"GET", "/unknown", 404}} {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != test.code {
			t.Fatalf("%s %s: %d", test.method, test.path, response.Code)
		}
	}
}

func TestUnavailableCoreIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.sock")
	if runtime.GOOS == "windows" {
		path = fmt.Sprintf(`\\.\pipe\hopnode-poc-absent-%d-%d`, os.Getpid(), time.Now().UnixNano())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := query(ctx, path); err == nil {
		t.Fatal("unavailable Core must not return fixture data")
	}
}
