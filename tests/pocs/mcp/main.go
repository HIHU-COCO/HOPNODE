// Purpose: Isolated P03 MCP stdio protocol fixture; no production remote operations.
// Creator: HIHU-COCO
// Created: 2026-10-19
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type probeInput struct {
	DeviceID string `json:"device_id" jsonschema:"Explicit POC target, must equal poc-local"`
	Action   string `json:"action" jsonschema:"echo, fail, wait, or status"`
	Message  string `json:"message,omitempty" jsonschema:"Non-sensitive text to echo"`
}

type probeOutput struct {
	DeviceID    string `json:"device_id"`
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Active      int64  `json:"active"`
	Canceled    int64  `json:"canceled"`
}

func newServer() *mcp.Server {
	var active, canceled, sequence atomic.Int64
	openWorld, destructive := false, false
	server := mcp.NewServer(&mcp.Implementation{Name: "hopnode", Version: "0.0.0-poc.3"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "poc_probe",
		Description: "P03 protocol fixture only. Requires device_id=poc-local. Echo returns text; fail returns a tool error; wait blocks until canceled; status exposes fixture cancellation counters. No SSH, filesystem, or production devices.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: &destructive, OpenWorldHint: &openWorld},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input probeInput) (*mcp.CallToolResult, probeOutput, error) {
		if input.DeviceID != "poc-local" {
			return nil, probeOutput{}, errors.New("POC_DEVICE_REQUIRED: device_id must equal poc-local")
		}
		output := probeOutput{DeviceID: input.DeviceID, OperationID: "poc-op-" + strconv.FormatInt(sequence.Add(1), 10), Status: "succeeded"}
		switch input.Action {
		case "echo":
			output.Message = input.Message
		case "fail":
			return nil, probeOutput{}, errors.New("POC_EXPECTED_FAILURE: requested fixture error")
		case "wait":
			active.Add(1)
			defer active.Add(-1)
			<-ctx.Done()
			canceled.Add(1)
			log.New(os.Stderr, "hopnode-poc: ", 0).Print("P03_WAIT_CANCELED")
			return nil, probeOutput{}, ctx.Err()
		case "status":
			output.Active, output.Canceled = active.Load(), canceled.Load()
		default:
			return nil, probeOutput{}, errors.New("POC_ACTION_INVALID: unknown action")
		}
		return nil, output, nil
	})
	return server
}

func run() error {
	// Logs use stderr; stdout is reserved exclusively for newline JSON-RPC frames.
	log.New(os.Stderr, "hopnode-poc: ", 0).Print("stdio ready")
	return newServer().Run(context.Background(), &mcp.StdioTransport{})
}

func main() {
	if err := run(); err != nil {
		log.New(os.Stderr, "hopnode-poc: ", 0).Print(err)
		os.Exit(1)
	}
}
