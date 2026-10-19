// Purpose: Offline DRAFT Directory OpenAPI fragments and IPC event examples.
// Creator: HIHU-COCO
// Created: 2026-10-19
package contractsdraft

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestDirectoryAndEventExamples(t *testing.T) {
	var cases []struct {
		Name       string          `json:"name"`
		Schema     string          `json:"schema"`
		Definition string          `json:"definition"`
		Valid      bool            `json:"valid"`
		Value      json.RawMessage `json:"value"`
	}
	readJSON(t, "examples/directory-events-cases.json", &cases)
	if len(cases) == 0 {
		t.Fatal("no Directory/event candidate examples")
	}
	positive := 0
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
		if example.Valid {
			positive++
		}
	}
	t.Logf("DRAFT Directory/events: %d positive, %d negative; no HTTP/IPC execution", positive, len(cases)-positive)
}

func TestDirectoryMetadataPatchRejectsProtectedFields(t *testing.T) {
	schema := resolve(t, "directory.schema.json", "UpdateDeviceRequest")
	for name, value := range map[string]any{
		"device_id": "draft-other-device", "source": "managed",
		"agent_enabled": true, "identity_status": "revoked",
		"authorization_version": 1, "device_key_id": "draft-key-001",
	} {
		t.Run(name, func(t *testing.T) {
			patch := map[string]any{"note": "DRAFT allowed note", name: value}
			if err := schema.Validate(patch); err == nil {
				t.Fatal("metadata patch accepted a protected field")
			}
		})
	}
}

// This checks this artifact's documented fragment coverage and resolves each
// JSON Schema reference. It is not a complete OpenAPI specification validator.
func TestDirectoryOpenAPICandidateInventoryAndReferences(t *testing.T) {
	var document struct {
		OpenAPI string `json:"openapi"`
		Info    struct {
			Title   string `json:"title"`
			Version string `json:"version"`
		} `json:"info"`
		Draft              bool   `json:"x-hopnode-draft"`
		AuthenticationTODO string `json:"x-hopnode-authentication-todo"`
		Servers            []struct {
			URL string `json:"url"`
		} `json:"servers"`
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	readJSON(t, "schemas/directory.openapi.json", &document)
	if document.OpenAPI != "3.1.0" || !document.Draft || !strings.HasPrefix(document.Info.Title, "DRAFT ") ||
		!strings.Contains(document.Info.Version, "DRAFT") || document.AuthenticationTODO == "" ||
		len(document.Servers) != 1 || document.Servers[0].URL != "https://directory.invalid" {
		t.Fatal("OpenAPI candidate must retain DRAFT and unresolved deployment boundaries")
	}
	// Independent Design 7.3 inventory: seven paths, eight operations.
	expected := map[string][]string{
		"/api/v1/enrollments":            {"post"},
		"/api/v1/devices/register":       {"post"},
		"/api/v1/devices/{id}/heartbeat": {"post"},
		"/api/v1/devices":                {"get"},
		"/api/v1/devices/{id}":           {"get", "patch"},
		"/api/v1/devices/{id}/disable":   {"post"},
		"/healthz":                       {"get"},
	}
	if len(document.Paths) != len(expected) {
		t.Fatal("Directory path inventory differs from Design 7.3")
	}
	fragments, inventoryOnly := 0, 0
	for path, methods := range expected {
		item, ok := document.Paths[path]
		if !ok {
			t.Fatalf("missing Design path %s", path)
		}
		for _, method := range methods {
			var operation struct {
				State     string                     `json:"x-hopnode-draft-state"`
				TODO      string                     `json:"x-hopnode-todo"`
				Responses map[string]json.RawMessage `json:"responses"`
			}
			if err := json.Unmarshal(item[method], &operation); err != nil {
				t.Fatalf("missing or invalid %s %s: %v", method, path, err)
			}
			if len(operation.Responses) == 0 {
				t.Fatal("OpenAPI operation requires responses")
			}
			switch operation.State {
			case "schema-fragment":
				fragments++
				if len(operation.Responses["200"]) == 0 {
					t.Fatal("documented response fragment missing")
				}
			case "inventory-only":
				inventoryOnly++
				if operation.TODO == "" {
					t.Fatal("unimplemented contract must retain its TODO")
				}
			default:
				t.Fatal("operation cannot silently appear complete")
			}
		}
		if strings.Contains(path, "{id}") {
			var parameters []struct {
				Name     string `json:"name"`
				In       string `json:"in"`
				Required bool   `json:"required"`
			}
			if err := json.Unmarshal(item["parameters"], &parameters); err != nil || len(parameters) != 1 ||
				parameters[0].Name != "id" || parameters[0].In != "path" || !parameters[0].Required {
				t.Fatal("target path parameter must be explicit and required")
			}
		}
		for key := range item {
			if key == "parameters" {
				continue
			}
			allowed := false
			for _, method := range methods {
				allowed = allowed || key == method
			}
			if !allowed {
				t.Fatal("undocumented Directory operation")
			}
		}
	}
	if fragments != 3 || inventoryOnly != 5 {
		t.Fatal("Directory coverage must remain three body fragments and five TODO operations")
	}
	var raw any
	readJSON(t, "schemas/directory.openapi.json", &raw)
	base, err := url.Parse(schemaBase + "directory.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	references := 0
	var inspect func(any)
	inspect = func(value any) {
		switch node := value.(type) {
		case map[string]any:
			for key, child := range node {
				if key == "$ref" {
					address, ok := child.(string)
					if !ok {
						t.Fatal("invalid schema reference")
					}
					reference, err := url.Parse(address)
					if err != nil {
						t.Fatal(err)
					}
					target := base.ResolveReference(reference)
					// The shared loader rejects all unlisted schemas and I/O fallback.
					wrapper := &jsonschema.Schema{Ref: target.String()}
					if _, err := wrapper.Resolve(&jsonschema.ResolveOptions{Loader: loadSchema, ValidateDefaults: true}); err != nil {
						t.Fatalf("unresolved OpenAPI body/parameter schema: %v", err)
					}
					references++
				} else {
					inspect(child)
				}
			}
		case []any:
			for _, child := range node {
				inspect(child)
			}
		}
	}
	inspect(raw)
	if references == 0 {
		t.Fatal("no machine-readable schema fragments")
	}
	t.Logf("DRAFT OpenAPI: 8 operation inventory entries, 3 body fragments, 5 TODO entries, %d resolved schema references", references)
}
