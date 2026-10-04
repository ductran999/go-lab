package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// feed runs input lines through a server with one echo tool.
func feed(t *testing.T, lines ...string) []string {
	t.Helper()

	srv := New("test", "0.0.1")
	srv.Add(Tool{Name: "echo", Description: "echo", Schema: map[string]any{}}, func(args map[string]any) (string, error) {
		return StrArg(args, "in", "empty"), nil
	})

	var out bytes.Buffer

	srv.Serve(strings.NewReader(strings.Join(lines, "\n")+"\n"), &out)

	var resps []string

	for line := range strings.Lines(strings.TrimSpace(out.String())) {
		resps = append(resps, line)
	}

	return resps
}

func get(t *testing.T, raw string) map[string]any {
	t.Helper()

	var m map[string]any

	err := json.Unmarshal([]byte(raw), &m)
	if err != nil {
		t.Fatalf("response not JSON: %v\n%s", err, raw)
	}

	return m
}

func TestInitializeEchoesProtocol(t *testing.T) {
	t.Parallel()

	resps := feed(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2099-01-01"}}`)
	if len(resps) != 1 {
		t.Fatalf("want 1 response, got %d", len(resps))
	}

	result := get(t, resps[0])["result"].(map[string]any)
	if result["protocolVersion"] != "2099-01-01" {
		t.Fatalf("protocol = %v, want client echo", result["protocolVersion"])
	}
}

func TestToolsListAndCall(t *testing.T) {
	t.Parallel()

	resps := feed(t,
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"in":"hi"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":3,"method":"nope"}`,
	)
	if len(resps) != 3 {
		t.Fatalf("want 3 responses (notification is silent), got %d", len(resps))
	}

	tools := get(t, resps[0])["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("tools = %d, want 1", len(tools))
	}

	content := get(t, resps[1])["result"].(map[string]any)["content"].([]any)
	if content[0].(map[string]any)["text"] != "hi" {
		t.Fatalf("call result = %v, want hi", content)
	}

	errObj := get(t, resps[2])["error"].(map[string]any)
	if errObj["code"].(float64) != CodeMethodNotFound {
		t.Fatalf("code = %v, want method-not-found", errObj["code"])
	}
}

func TestUnknownToolIsProtocolError(t *testing.T) {
	t.Parallel()

	resps := feed(t, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ghost"}}`)
	if len(resps) != 1 {
		t.Fatalf("want 1 response, got %d", len(resps))
	}

	m := get(t, resps[0])
	if m["result"] != nil {
		t.Fatalf("unknown tool must be a protocol error, not a result: %s", resps[0])
	}

	if m["error"].(map[string]any)["code"].(float64) != CodeMethodNotFound {
		t.Fatalf("want method-not-found: %s", resps[0])
	}
}

func TestIntArgClamps(t *testing.T) {
	t.Parallel()

	args := map[string]any{"n": 99999.0, "neg": -5.0}

	if got := IntArg(args, "n", 1, 1, 100); got != 100 {
		t.Fatalf("clamp high = %d, want 100", got)
	}

	if got := IntArg(args, "neg", 1, 1, 100); got != 1 {
		t.Fatalf("clamp low = %d, want 1", got)
	}

	if got := IntArg(args, "missing", 7, 1, 100); got != 7 {
		t.Fatalf("default = %d, want 7", got)
	}
}
