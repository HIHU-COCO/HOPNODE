// Purpose: Offline DRAFT schema resolution and positive/negative contract examples.
// Creator: HIHU-COCO
// Created: 2026-10-19
package contractsdraft

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed schemas/*.json examples/cases.json examples/directory-events-cases.json examples/valid/tool-calls.json
var documents embed.FS

const schemaBase = "https://hopnode.invalid/draft/"

func loadSchema(uri *url.URL) (*jsonschema.Schema, error) {
	// Only embedded, exact identifiers are allowed. There is no HTTP/file fallback.
	for _, name := range []string{"core-api.schema.json", "errors.schema.json", "mcp-tools.schema.json", "directory.schema.json", "ipc-events.schema.json"} {
		if uri.String() != schemaBase+name {
			continue
		}
		data, err := documents.ReadFile("schemas/" + name)
		if err != nil {
			return nil, err
		}
		var schema jsonschema.Schema
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, err
		}
		return &schema, nil
	}
	return nil, fmt.Errorf("schema reference outside embedded DRAFT allowlist")
}

func resolve(t *testing.T, file, definition string) *jsonschema.Resolved {
	t.Helper()
	uri, err := url.Parse(schemaBase + file)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := loadSchema(uri)
	if err != nil {
		t.Fatal(err)
	}
	if definition != "" {
		schema = &jsonschema.Schema{Ref: schemaBase + file + "#/$defs/" + definition}
	}
	resolved, err := schema.Resolve(&jsonschema.ResolveOptions{Loader: loadSchema, ValidateDefaults: true})
	if err != nil {
		t.Fatalf("DRAFT schema %s/%s: %v", file, definition, err)
	}
	return resolved
}

func readJSON(t *testing.T, file string, destination any) {
	t.Helper()
	data, err := documents.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		t.Fatal(err)
	}
}

func TestSchemasResolveOffline(t *testing.T) {
	for _, name := range []string{"core-api.schema.json", "errors.schema.json", "mcp-tools.schema.json", "directory.schema.json", "ipc-events.schema.json"} {
		t.Run(name, func(t *testing.T) {
			resolved := resolve(t, name, "")
			if !strings.HasPrefix(resolved.Schema().Title, "DRAFT ") || resolved.Schema().Schema != "https://json-schema.org/draft/2020-12/schema" {
				t.Fatal("schema must declare DRAFT status and supported dialect")
			}
		})
	}
	for _, address := range []string{"https://example.invalid/other.json", "file:///etc/passwd", schemaBase + "../other.json", schemaBase + "errors.schema.json?token=SYNTHETIC"} {
		uri, err := url.Parse(address)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := loadSchema(uri); err == nil {
			t.Fatal("unexpected external schema load")
		}
	}
}

func TestPositiveAndNegativeExamples(t *testing.T) {
	var cases []struct {
		Name       string          `json:"name"`
		Schema     string          `json:"schema"`
		Definition string          `json:"definition"`
		Valid      bool            `json:"valid"`
		Value      json.RawMessage `json:"value"`
	}
	readJSON(t, "examples/cases.json", &cases)
	if len(cases) == 0 {
		t.Fatal("no DRAFT examples")
	}
	for _, example := range cases {
		t.Run(example.Name, func(t *testing.T) {
			var value any
			if err := json.Unmarshal(example.Value, &value); err != nil {
				t.Fatal(err)
			}
			err := resolve(t, example.Schema, example.Definition).Validate(value)
			if example.Valid && err != nil {
				t.Fatalf("positive DRAFT example rejected: %v", err)
			}
			if !example.Valid && err == nil {
				t.Fatal("negative DRAFT example accepted")
			}
		})
	}
	t.Logf("DRAFT examples checked: %d; no runtime/platform claims", len(cases))
}

// This list is the Design 9.3 inventory, independent of the schema's oneOf.
const expectedTools = "list_devices get_device connect_device disconnect_device list_files search_files read_file write_file apply_patch make_directory move_path copy_path delete_path upload_file download_file exec_command open_terminal read_terminal write_terminal close_terminal list_tasks get_task cancel_task list_backups restore_backup get_operation_history"

func toolExamples(t *testing.T) map[string]map[string]any {
	t.Helper()
	var calls []map[string]any
	readJSON(t, "examples/valid/tool-calls.json", &calls)
	byName := map[string]map[string]any{}
	for _, call := range calls {
		name, ok := call["name"].(string)
		if !ok || byName[name] != nil {
			t.Fatal("invalid or duplicate DRAFT tool example")
		}
		byName[name] = call
	}
	return byName
}

func clone(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestAll26ToolsAndExplicitTargetRejections(t *testing.T) {
	schema := resolve(t, "mcp-tools.schema.json", "")
	examples := toolExamples(t)
	expected := strings.Fields(expectedTools)
	actual := make([]string, 0, len(examples))
	for name := range examples {
		actual = append(actual, name)
	}
	sort.Strings(actual)
	sort.Strings(expected)
	if strings.Join(actual, " ") != strings.Join(expected, " ") || len(expected) != 26 {
		t.Fatal("DRAFT tool examples differ from Design 9.3")
	}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			call := examples[name]
			if err := schema.Validate(call); err != nil {
				t.Fatalf("valid DRAFT tool call rejected: %v", err)
			}
			if name == "list_devices" {
				return // Global discovery has no target; all other candidates do.
			}
			for _, variant := range []string{"missing", "wrong_type", "empty", "camel_case"} {
				invalid := clone(t, call)
				args := invalid["arguments"].(map[string]any)
				switch variant {
				case "missing":
					delete(args, "device_id")
				case "wrong_type":
					args["device_id"] = float64(42)
				case "empty":
					args["device_id"] = ""
				case "camel_case":
					args["deviceId"] = args["device_id"]
					delete(args, "device_id")
				}
				if err := schema.Validate(invalid); err == nil {
					t.Fatalf("accepted %s target", variant)
				}
			}
		})
	}
	if err := schema.Validate(map[string]any{"name": "poc_probe", "arguments": map[string]any{"device_id": "poc-local"}}); err == nil {
		t.Fatal("isolated P03 fixture must not become a production candidate")
	}
	t.Log("DRAFT: 26 valid tool examples, 25 explicit targets, 100 invalid target variants rejected")
}

func TestReplaySensitiveCandidatesRequireOperationID(t *testing.T) {
	schema := resolve(t, "mcp-tools.schema.json", "")
	examples := toolExamples(t)
	for _, name := range strings.Fields("write_file apply_patch make_directory move_path copy_path delete_path upload_file download_file exec_command write_terminal cancel_task restore_backup") {
		t.Run(name, func(t *testing.T) {
			invalid := clone(t, examples[name])
			delete(invalid["arguments"].(map[string]any), "operation_id")
			if err := schema.Validate(invalid); err == nil {
				t.Fatal("replay-sensitive DRAFT input missing operation_id accepted")
			}
		})
	}
}

func TestDesignRequiredBusinessErrors(t *testing.T) {
	schema := resolve(t, "errors.schema.json", "")
	for _, code := range strings.Fields("DEVICE_OFFLINE NETWORK_UNREACHABLE SSH_NOT_READY SSH_AUTH_FAILED SSH_HOST_KEY_CHANGED PERMISSION_DENIED MCP_NOT_CONNECTED TASK_STATUS_UNKNOWN FILE_CONFLICT SERVICE_NOT_RUNNING") {
		t.Run(code, func(t *testing.T) {
			value := map[string]any{"code": code, "message": "DRAFT diagnostic", "device_id": "draft-device-001", "retryable": false, "details": map[string]any{}}
			if err := schema.Validate(value); err != nil {
				t.Fatalf("Design 10.1 required error rejected: %v", err)
			}
		})
	}
}
