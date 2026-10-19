// Purpose: relay preserves frames and omits non-handshake payloads. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestForwardAndHandshakeOnlyTrace(t *testing.T) {
	var forwarded, trace bytes.Buffer
	r := &recorder{trace: &trace}
	input := "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2025-11-25\",\"clientInfo\":{\"name\":\"test\",\"version\":\"1\"}}}\r\n" +
		"{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"secret_fixture\":\"do-not-record\"}}\n"
	if err := forward(&forwarded, strings.NewReader(input), func(line []byte) { r.observe(line, true) }); err != nil {
		t.Fatal(err)
	}
	if forwarded.String() != input {
		t.Fatal("relay altered protocol bytes")
	}
	response := "{\"id\":3,\"result\":{\"protocolVersion\":\"2025-11-25\",\"secret_fixture\":\"do-not-record\"}}\r\n{\"message\":\"中文 final partial frame\"}"
	forwarded.Reset()
	if err := forward(&forwarded, strings.NewReader(response), func(line []byte) { r.observe(line, false) }); err != nil || forwarded.String() != response {
		t.Fatal("relay altered server bytes or partial EOF frame")
	}
	if !strings.Contains(trace.String(), "initialize_request") || !strings.Contains(trace.String(), "initialize_response") || strings.Contains(trace.String(), "do-not-record") || strings.Contains(trace.String(), "secret_fixture") {
		t.Fatal("incorrect handshake-only evidence")
	}
}

func TestCancellationTraceHasOnlyEventNames(t *testing.T) {
	var trace bytes.Buffer
	r := &recorder{trace: &trace, cancellation: true}
	r.observe([]byte(`{"id":9,"method":"tools/call","params":{"arguments":{"action":"wait","message":"do-not-record"}}}`), true)
	r.observe([]byte(`{"method":"notifications/cancelled","params":{"requestId":8,"reason":"do-not-record"}}`), true)
	r.observe([]byte(`{"method":"notifications/cancelled","params":{"requestId":9,"reason":"do-not-record"}}`), true)
	if got, want := trace.String(), "{\"event\":\"wait_call\"}\n{\"event\":\"wait_cancel_notification\"}\n"; got != want {
		t.Fatalf("unexpected allowlisted trace: %q", got)
	}
}
