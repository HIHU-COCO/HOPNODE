// Purpose: P03 real subprocess stdio, negotiation, error, cancellation, and shutdown checks.
// Creator: HIHU-COCO
// Created: 2026-10-19
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServerProcess(t *testing.T) {
	if os.Getenv("HOPNODE_POC_SERVER_PROCESS") != "1" {
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0) // Suppress the Go test runner's PASS line on protocol stdout.
}

func serverCommand(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestServerProcess$")
	cmd.Env = append(os.Environ(), "HOPNODE_POC_SERVER_PROCESS=1")
	return cmd
}

type lockedBuffer struct {
	mu sync.Mutex
	bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}

func (b *lockedBuffer) text() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.String()
}

func TestSDKSubprocess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := serverCommand(ctx)
	var diagnostics lockedBuffer
	cmd.Stderr = &diagnostics
	client := mcp.NewClient(&mcp.Implementation{Name: "hopnode-poc-test", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	init := session.InitializeResult()
	if init.ProtocolVersion != "2026-07-28" || init.ServerInfo.Name != "hopnode" {
		t.Fatalf("unexpected negotiated implementation: %+v", init)
	}
	t.Logf("SDK client negotiated %s with %s", init.ProtocolVersion, init.ServerInfo.Name)
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 1 || tools.Tools[0].Name != "poc_probe" {
		t.Fatalf("tools=%+v err=%v", tools, err)
	}
	call := func(action string) (*mcp.CallToolResult, error) {
		return session.CallTool(ctx, &mcp.CallToolParams{Name: "poc_probe", Arguments: probeInput{DeviceID: "poc-local", Action: action, Message: "中文 stdio"}})
	}
	result, err := call("echo")
	if err != nil || result.IsError {
		t.Fatalf("echo=%+v err=%v", result, err)
	}
	var output probeOutput
	decodeOutput(t, result, &output)
	if output.DeviceID != "poc-local" || output.Message != "中文 stdio" || output.OperationID == "" {
		t.Fatalf("unexpected output: %+v", output)
	}
	result, err = call("fail")
	if err != nil || !result.IsError || !strings.Contains(resultText(result), "POC_EXPECTED_FAILURE") {
		t.Fatalf("expected tool error, result=%+v err=%v", result, err)
	}
	for _, args := range []any{map[string]any{"action": "echo"}, map[string]any{"device_id": 123, "action": "echo"}, probeInput{DeviceID: "other-device", Action: "echo"}} {
		result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "poc_probe", Arguments: args})
		if err == nil && !result.IsError {
			t.Fatalf("invalid target/schema accepted: %#v", args)
		}
	}
	// Cancel only after a separate tool call observes that the remote handler is running.
	// This proves notifications/cancelled reaches the server, not merely a local timeout.
	waitCtx, stopWait := context.WithCancel(ctx)
	completed := make(chan error, 1)
	go func() {
		_, err := session.CallTool(waitCtx, &mcp.CallToolParams{Name: "poc_probe", Arguments: probeInput{DeviceID: "poc-local", Action: "wait"}})
		completed <- err
	}()
	defer stopWait()
	poll := func(wantActive, wantCanceled int64) {
		t.Helper()
		for {
			result, err := call("status")
			if err != nil {
				t.Fatal(err)
			}
			decodeOutput(t, result, &output)
			if output.Active == wantActive && output.Canceled == wantCanceled {
				return
			}
			select {
			case <-ctx.Done():
				t.Fatalf("handler state %+v: %v", output, ctx.Err())
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	poll(1, 0)
	stopWait()
	if err := <-completed; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled call: %v", err)
	}
	poll(0, 1)
	if !strings.Contains(diagnostics.text(), "stdio ready") {
		t.Fatal("missing stderr startup diagnostic")
	}
	if err := session.Close(); err != nil {
		t.Fatalf("graceful stdio EOF shutdown: %v", err)
	}
	t.Log("tools/list, structured echo, errors, explicit target, remote cancellation, stderr, EOF shutdown PASS")
}

func resultText(result *mcp.CallToolResult) string {
	var text strings.Builder
	for _, item := range result.Content {
		if item, ok := item.(*mcp.TextContent); ok {
			text.WriteString(item.Text)
		}
	}
	return text.String()
}

func decodeOutput(t *testing.T, result *mcp.CallToolResult, target *probeOutput) {
	t.Helper()
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyWireAndPureStdout(t *testing.T) {
	for _, version := range []string{"2024-11-05", "2025-03-26", "2025-06-18", "2025-11-25"} {
		t.Run(version, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := serverCommand(ctx)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			var diagnostics bytes.Buffer
			cmd.Stderr = &diagnostics
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer cmd.Process.Kill()
			reader := bufio.NewReader(stdout)
			send := func(value any) {
				t.Helper()
				if err := json.NewEncoder(stdin).Encode(value); err != nil {
					t.Fatal(err)
				}
			}
			receive := func() map[string]json.RawMessage {
				t.Helper()
				line, err := reader.ReadBytes('\n')
				if err != nil {
					t.Fatal(err)
				}
				var frame map[string]json.RawMessage
				if err := json.Unmarshal(line, &frame); err != nil || string(frame["jsonrpc"]) != `"2.0"` {
					t.Fatalf("stdout contains non-protocol content: %q (%v)", line, err)
				}
				return frame
			}
			send(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": version, "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "wire-test", "version": "1"}}})
			frame := receive()
			var initialization struct {
				ProtocolVersion string `json:"protocolVersion"`
			}
			if err := json.Unmarshal(frame["result"], &initialization); err != nil || initialization.ProtocolVersion != version {
				t.Fatalf("legacy negotiation result=%s err=%v", frame["result"], err)
			}
			send(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
			send(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": map[string]any{}})
			frame = receive()
			if !bytes.Contains(frame["result"], []byte("poc_probe")) {
				t.Fatalf("missing tool: %s", frame["result"])
			}
			send(map[string]any{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": map[string]any{"name": "poc_probe", "arguments": probeInput{DeviceID: "poc-local", Action: "echo", Message: version}}})
			frame = receive()
			if !bytes.Contains(frame["result"], []byte(version)) {
				t.Fatalf("missing echo: %s", frame["result"])
			}
			// Malformed params produce a protocol error, distinct from tool isError.
			send(map[string]any{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": "invalid-params"})
			frame = receive()
			var rpcError struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(frame["error"], &rpcError); err != nil || rpcError.Code != -32602 {
				t.Fatalf("invalid params must be -32602, got %s (%v)", frame["error"], err)
			}
			if err := stdin.Close(); err != nil {
				t.Fatal(err)
			}
			remainder, err := io.ReadAll(reader)
			if err != nil || len(remainder) != 0 {
				t.Fatalf("unexpected trailing stdout %q: %v", remainder, err)
			}
			if err := cmd.Wait(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(diagnostics.String(), "stdio ready") {
				t.Fatal("expected log on stderr")
			}
			t.Logf("legacy %s initialization, tool execution, -32602, stdout JSONL, shutdown PASS", version)
		})
	}
}
