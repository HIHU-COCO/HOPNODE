//go:build linux

// Purpose: prove audit PID/UID comes from the real UDS client process. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditUsesKernelClientIdentity(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "core.sock")
	listener, err := listen(path)
	if err != nil {
		t.Fatal(err)
	}
	audit, err := os.OpenFile(filepath.Join(dir, "audit.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer audit.Close()
	status, err := newStatus()
	if err != nil {
		t.Fatal(err)
	}
	server := newServer(status, audit)
	defer server.Close()
	go func() { _ = server.Serve(listener) }()
	if _, err := query(context.Background(), path); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(audit.Name())
	if err != nil {
		t.Fatal(err)
	}
	var event struct {
		peerIdentity
		Path       string `json:"path"`
		HTTPStatus int    `json:"http_status"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		t.Fatal(err)
	}
	if !event.Verified || event.PID != os.Getpid() || event.UID != os.Getuid() || event.Path != "/v0/status" || event.HTTPStatus != 200 {
		t.Fatalf("unexpected kernel peer evidence: %+v", event)
	}
}

func TestAuditNeverOverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	auditPath := filepath.Join(dir, "audit.jsonl")
	if err := os.WriteFile(auditPath, []byte("keep existing evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOPNODE_POC_AUDIT_FILE", auditPath)
	if err := serve(context.Background(), filepath.Join(dir, "core.sock")); err == nil {
		t.Fatal("existing audit must be rejected")
	}
	body, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "keep existing evidence" {
		t.Fatal("existing evidence was overwritten")
	}
}
