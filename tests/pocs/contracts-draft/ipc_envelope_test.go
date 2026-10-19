// Purpose: Offline DRAFT IPC response mapping and bounded request/response checks.
// Creator: HIHU-COCO
// Created: 2026-10-19
package contractsdraft

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/jsonschema-go/jsonschema"
)

// These limits apply to the offline checker, not an implemented IPC transport.
const envelopeBodyLimit = 64 * 1024
const envelopeDepthLimit = 32

//go:embed schemas/ipc-envelope.schema.json envelope-cases/*.json
var envelopeDocuments embed.FS

func envelopeLoader(uri *url.URL) (*jsonschema.Schema, error) {
	if uri.String() != schemaBase+"ipc-envelope.schema.json" {
		return loadSchema(uri)
	}
	data, err := envelopeDocuments.ReadFile("schemas/ipc-envelope.schema.json")
	if err != nil {
		return nil, err
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

func envelopeResolve(t *testing.T, definition string) *jsonschema.Resolved {
	t.Helper()
	ref := schemaBase + "ipc-envelope.schema.json"
	if definition != "" {
		ref += "#/$defs/" + definition
	}
	schema := &jsonschema.Schema{Ref: ref}
	resolved, err := schema.Resolve(&jsonschema.ResolveOptions{Loader: envelopeLoader, ValidateDefaults: true})
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

// parseEnvelopeObject is a test-only, byte/depth-bounded JSON checker. It rejects
// duplicate keys and trailing values instead of silently selecting a different
// request identity. It does not read a pipe, socket, length header or timer.
func parseEnvelopeObject(data []byte) (map[string]any, error) {
	if len(data) == 0 || len(data) > envelopeBodyLimit || !utf8.Valid(data) {
		return nil, fmt.Errorf("invalid candidate JSON body size or UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var consume func(int) error
	consume = func(depth int) error {
		if depth > envelopeDepthLimit {
			return fmt.Errorf("candidate JSON depth exceeded")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, container := token.(json.Delim)
		if !container {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] {
					return fmt.Errorf("duplicate or invalid JSON key")
				}
				seen[name] = true
				if err := consume(depth + 1); err != nil {
					return err
				}
			}
		case '[':
			for decoder.More() {
				if err := consume(depth + 1); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter")
		}
		_, err = decoder.Token()
		return err
	}
	if err := consume(0); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing JSON content")
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil || value == nil {
		return nil, fmt.Errorf("candidate JSON body must be an object")
	}
	return value, nil
}

type envelopePairChecker struct {
	request  *jsonschema.Resolved
	response *jsonschema.Resolved
	id       *jsonschema.Resolved
}

func newEnvelopePairChecker(t *testing.T) envelopePairChecker {
	t.Helper()
	return envelopePairChecker{
		request:  resolve(t, "core-api.schema.json", "CoreRequest"),
		response: envelopeResolve(t, "Response"),
		id:       resolve(t, "core-api.schema.json", "ID"),
	}
}

func (checker envelopePairChecker) check(requestJSON, responseJSON []byte) error {
	response, err := parseEnvelopeObject(responseJSON)
	if err != nil {
		return err
	}
	if err := checker.response.Validate(response); err != nil {
		return fmt.Errorf("response schema: %w", err)
	}
	request, requestErr := parseEnvelopeObject(requestJSON)
	if response["outcome"] == "protocol_error" {
		code := response["error"].(map[string]any)["code"]
		if requestErr != nil && code != "PARSE_ERROR" {
			return fmt.Errorf("unparseable candidate body requires parse error")
		}
		if requestErr == nil && code == "PARSE_ERROR" {
			return fmt.Errorf("parsed request cannot claim parse failure")
		}
		if requestErr == nil && checker.request.Validate(request) == nil && code != "IPC_VERSION_UNSUPPORTED" {
			return fmt.Errorf("schema-valid request cannot claim malformed request or unknown method")
		}
		// A null association is permitted only when the bounded parser cannot
		// identify that metadata. A recognized ID/method must be echoed even
		// when the method or its parameters are invalid.
		var requestID, method any
		if requestErr == nil {
			if candidate := request["request_id"]; checker.id.Validate(candidate) == nil {
				requestID = candidate
			}
			if candidate, ok := request["method"].(string); ok && len(candidate) > 0 && utf8.RuneCountInString(candidate) <= 128 {
				method = candidate
			}
		}
		if response["request_id"] != requestID || response["method"] != method {
			return fmt.Errorf("protocol response lost or changed recognizable association")
		}
		return nil
	}
	if requestErr != nil {
		return requestErr
	}
	if err := checker.request.Validate(request); err != nil {
		return fmt.Errorf("non-protocol response requires a valid request: %w", err)
	}
	if response["request_id"] != request["request_id"] || response["method"] != request["method"] {
		return fmt.Errorf("response request_id or method differs from request")
	}
	params := request["params"].(map[string]any)
	if response["outcome"] == "business_error" {
		businessError := response["error"].(map[string]any)
		if businessError["device_id"] != params["device_id"] {
			return fmt.Errorf("business error target differs from request")
		}
		for _, key := range []string{"operation_id", "task_id"} {
			if expected, present := params[key]; present && response[key] != expected {
				return fmt.Errorf("business error lost request %s correlation", key)
			}
			if _, present := params[key]; !present && response[key] != nil {
				return fmt.Errorf("business error invented request %s correlation", key)
			}
		}
		return nil
	}
	data := response["data"].(map[string]any)
	for _, key := range []string{"device_id", "task_id"} {
		if expected, present := params[key]; present && data[key] != expected {
			return fmt.Errorf("result %s differs from request", key)
		}
	}
	// CancelTask.operation_id remains ambiguous in the existing DRAFT. Do not
	// equate it to Task.operation_id or invent a production meaning here.
	if request["method"] == "GetOperation" || request["method"] == "CancelOperation" {
		if data["operation_id"] != params["operation_id"] {
			return fmt.Errorf("result operation_id differs from target operation")
		}
		if targetError, ok := data["error"].(map[string]any); ok && targetError["device_id"] != data["device_id"] {
			return fmt.Errorf("target operation error has another device_id")
		}
	}
	return nil
}

func TestIPCEnvelopeCandidateOfflineResolution(t *testing.T) {
	resolved := envelopeResolve(t, "")
	if resolved.Schema().Ref == "" {
		t.Fatal("missing candidate root reference")
	}
	for _, address := range []string{
		"https://example.invalid/ipc-envelope.schema.json", "file:///etc/passwd",
		schemaBase + "ipc-envelope.schema.json?secret=SYNTHETIC", schemaBase + "other.json",
	} {
		uri, err := url.Parse(address)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := envelopeLoader(uri); err == nil {
			t.Fatal("candidate loader accepted external or non-allowlisted input")
		}
	}
}

func TestIPCEnvelopeResponseAndPairExamples(t *testing.T) {
	data, err := envelopeDocuments.ReadFile("envelope-cases/pairs.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name          string          `json:"name"`
		Valid         bool            `json:"valid"`
		ResponseValid bool            `json:"response_valid"`
		Request       json.RawMessage `json:"request"`
		Response      json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	checker := newEnvelopePairChecker(t)
	positive, negative, schemaNegative := 0, 0, 0
	successfulMethods := map[string]bool{}
	for _, example := range cases {
		t.Run(example.Name, func(t *testing.T) {
			response, err := parseEnvelopeObject(example.Response)
			if err != nil {
				t.Fatal(err)
			}
			err = checker.response.Validate(response)
			if (err == nil) != example.ResponseValid {
				t.Fatalf("response schema valid=%v, want %v: %v", err == nil, example.ResponseValid, err)
			}
			err = checker.check(example.Request, example.Response)
			if (err == nil) != example.Valid {
				t.Fatalf("pair valid=%v, want %v: %v", err == nil, example.Valid, err)
			}
		})
		if example.Valid {
			positive++
			var response struct {
				Method  string `json:"method"`
				Outcome string `json:"outcome"`
			}
			if err := json.Unmarshal(example.Response, &response); err != nil {
				t.Fatal(err)
			}
			if response.Outcome == "success" {
				successfulMethods[response.Method] = true
			}
		} else {
			negative++
		}
		if !example.ResponseValid {
			schemaNegative++
		}
	}
	if positive < 6 || negative < 12 || schemaNegative < 6 {
		t.Fatal("candidate examples lost minimum positive/negative coverage")
	}
	// This inventory is independent of counts and the response schema. Losing
	// any one method's success example must fail even if other cases grow.
	for _, method := range strings.Fields("GetCapabilities ListDevices GetTask CancelTask GetOperation CancelOperation") {
		if !successfulMethods[method] {
			t.Fatalf("missing explicit success mapping example for %s", method)
		}
	}
	t.Logf("DRAFT envelope: %d valid pairs, %d rejected pairs (%d schema failures); no IPC transport", positive, negative, schemaNegative)
}

func TestIPCEnvelopeBoundedJSONAndUnparseableRequest(t *testing.T) {
	checker := newEnvelopePairChecker(t)
	protocolResponse := []byte(`{"message_type":"response","request_id":null,"method":null,"outcome":"protocol_error","data":null,"error":{"code":"PARSE_ERROR","message":"DRAFT malformed body"}}`)
	invalid := [][]byte{
		nil, []byte(`{"request_id":`), []byte(`{} {}`), []byte(`[]`), []byte(`null`),
		[]byte(`{"request_id":"draft-a","request_id":"draft-b"}`),
		[]byte(`{"params":{"device_id":"draft-a","device_id":"draft-b"}}`),
		[]byte("{\"bad\":\"\xff\"}"), bytes.Repeat([]byte(" "), envelopeBodyLimit+1),
		[]byte(`{"deep":` + strings.Repeat("[", envelopeDepthLimit+1) + "0" + strings.Repeat("]", envelopeDepthLimit+1) + "}"),
	}
	for index, body := range invalid {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			if _, err := parseEnvelopeObject(body); err == nil {
				t.Fatal("invalid or unbounded JSON accepted")
			}
			if err := checker.check(body, protocolResponse); err != nil {
				t.Fatalf("unrecognizable request must permit unassociated protocol error: %v", err)
			}
			if err := checker.check([]byte(`{}`), body); err == nil {
				t.Fatal("malformed response accepted")
			}
		})
	}
	// Exactly the candidate byte boundary is accepted as a JSON body; no claim
	// is made about a network reader enforcing that boundary before allocation.
	boundary := append([]byte(`{"pad":"`), bytes.Repeat([]byte("a"), envelopeBodyLimit-len(`{"pad":""}`))...)
	boundary = append(boundary, []byte(`"}`)...)
	if len(boundary) != envelopeBodyLimit {
		t.Fatal("incorrect boundary fixture")
	}
	if _, err := parseEnvelopeObject(boundary); err != nil {
		t.Fatal(err)
	}
}
