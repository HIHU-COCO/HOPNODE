//go:build windows

// Purpose: current-user-only Named Pipe transport. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"

	"github.com/Microsoft/go-winio"
)

func endpoint() (string, error) {
	if path := os.Getenv("HOPNODE_POC_ENDPOINT"); path != "" {
		return path, nil
	}
	name := os.Getenv("USERNAME")
	if name == "" {
		return "", fmt.Errorf("USERNAME is unavailable; set HOPNODE_POC_ENDPOINT")
	}
	return `\\.\pipe\hopnode-poc-` + name, nil
}

func listen(path string) (net.Listener, error) {
	current, err := user.Current()
	if err != nil {
		return nil, err
	}
	// Protected DACL grants only the current process user's SID full access.
	return winio.ListenPipe(path, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;" + current.Uid + ")", InputBufferSize: 4096, OutputBufferSize: 4096})
}

func dial(ctx context.Context, path string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, path)
}
