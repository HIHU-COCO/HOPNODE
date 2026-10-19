// Purpose: byte-preserving P03 stdio relay with allowlisted evidence. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"sync"
)

type recorder struct {
	mu           sync.Mutex
	request      string
	trace        io.Writer
	cancellation bool
	waitRequest  string
}

func (r *recorder) observe(line []byte, client bool) {
	if len(line) > 65536 {
		return
	}
	var frame struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
		Params struct {
			RequestID json.RawMessage `json:"requestId"`
			Arguments struct {
				Action string `json:"action"`
			} `json:"arguments"`
			Protocol string `json:"protocolVersion"`
			Client   struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"clientInfo"`
		} `json:"params"`
		Result struct {
			Protocol string `json:"protocolVersion"`
		} `json:"result"`
	}
	if json.Unmarshal(line, &frame) != nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if client && frame.Method == "initialize" {
		r.request = string(frame.ID)
		_ = json.NewEncoder(r.trace).Encode(map[string]string{"event": "initialize_request", "client_name": frame.Params.Client.Name, "client_version": frame.Params.Client.Version, "protocol_version": frame.Params.Protocol})
	} else if !client && r.request != "" && string(frame.ID) == r.request && frame.Result.Protocol != "" {
		_ = json.NewEncoder(r.trace).Encode(map[string]string{"event": "initialize_response", "protocol_version": frame.Result.Protocol})
	} else if r.cancellation && client && frame.Method == "tools/call" && frame.Params.Arguments.Action == "wait" {
		r.waitRequest = string(frame.ID)
		_ = json.NewEncoder(r.trace).Encode(map[string]string{"event": "wait_call"})
	} else if r.cancellation && client && frame.Method == "notifications/cancelled" && r.waitRequest != "" && string(frame.Params.RequestID) == r.waitRequest {
		_ = json.NewEncoder(r.trace).Encode(map[string]string{"event": "wait_cancel_notification"})
	}
}

func forward(destination io.Writer, source io.Reader, observe func([]byte)) error {
	reader := bufio.NewReader(source)
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) != 0 {
			observe(line)
			n, err := destination.Write(line)
			if err != nil {
				return err
			}
			if n != len(line) {
				return io.ErrShortWrite
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func run() error {
	target := flag.String("target", "", "isolated POC executable")
	tracePath := flag.String("trace", "", "new handshake-only trace file")
	cancellation := flag.Bool("trace-cancellation", false, "also record wait and matching cancellation event names only")
	flag.Parse()
	if *target == "" || *tracePath == "" || flag.NArg() != 0 {
		return errors.New("relay arguments required")
	}
	trace, err := os.OpenFile(*tracePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer trace.Close()
	r := &recorder{trace: trace, cancellation: *cancellation}
	command := exec.Command(*target)
	command.Stderr = os.Stderr
	input, err := command.StdinPipe()
	if err != nil {
		return err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = forward(input, os.Stdin, func(line []byte) { r.observe(line, true) }); _ = input.Close() }()
	if err := forward(os.Stdout, output, func(line []byte) { r.observe(line, false) }); err != nil {
		_ = input.Close()
		_ = command.Process.Kill()
		_ = command.Wait()
		return err
	}
	return command.Wait()
}

func main() {
	if run() != nil {
		_, _ = io.WriteString(os.Stderr, "P03_RELAY_FAILED\n")
		os.Exit(1)
	}
}
