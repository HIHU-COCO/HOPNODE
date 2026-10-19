//go:build !windows

// Purpose: private Unix Socket transport without deleting existing paths. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
)

func endpoint() (string, error) {
	if path := os.Getenv("HOPNODE_POC_ENDPOINT"); path != "" {
		return path, nil
	}
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "hopnode-poc.sock"), nil
	}
	return "", errors.New("set XDG_RUNTIME_DIR or HOPNODE_POC_ENDPOINT to a socket in a private user-owned directory")
}

func listen(path string) (net.Listener, error) {
	parent, err := os.Stat(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	owner, ok := parent.Sys().(*syscall.Stat_t)
	if !ok || owner.Uid != uint32(os.Getuid()) || parent.Mode().Perm()&0077 != 0 {
		return nil, errors.New("socket parent must be owned by current user with mode 0700")
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("refusing to replace existing or inaccessible socket path")
	}
	// Umask is set once at isolated fixture process startup, before serving requests.
	oldMask := syscall.Umask(0077)
	listener, err := net.Listen("unix", path)
	syscall.Umask(oldMask)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = listener.Close()
		return nil, err
	}
	return listener, nil
}

func dial(ctx context.Context, path string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "unix", path)
}
