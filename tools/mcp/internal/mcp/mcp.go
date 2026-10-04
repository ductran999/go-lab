// Package mcp speaks minimal MCP over stdio: initialize,
// tools/list, tools/call, ping. JSON-RPC 2.0, one object per line,
// logs on stderr — stdout stays pure protocol, or every client
// desyncs.
package mcp

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
)

// Method names we answer.
const (
	MethodInitialize = "initialize"
	MethodToolsList  = "tools/list"
	MethodToolsCall  = "tools/call"
	MethodPing       = "ping"
)

// DefaultProtocol echoes back when the client speaks unknown future;
// 2024-11-05 is the baseline every client understands.
const DefaultProtocol = "2024-11-05"

// JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternal       = -32603
)

// Request is one NDJSON line from the client.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is one NDJSON line back. Notifications (no ID) get none.
type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool describes one callable: name, blurb, JSON schema for args.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      any    `json:"inputSchema"` //nolint:tagliatelle // MCP wire names are camelCase by spec
}

// Handler runs a tool: args in, text out (or fail).
type Handler func(args map[string]any) (string, error)

// Server routes MCP methods to tool handlers.
type Server struct {
	info  map[string]string
	tools map[string]Tool
	run   map[string]Handler
}

// New names this server for the initialize handshake.
func New(name, version string) *Server {
	return &Server{
		info:  map[string]string{"name": name, "version": version},
		tools: map[string]Tool{},
		run:   map[string]Handler{},
	}
}

// Add registers one tool (definition + handler together, or they drift).
func (s *Server) Add(tool Tool, h Handler) {
	s.tools[tool.Name] = tool
	s.run[tool.Name] = h
}

// Serve reads requests until EOF or the first fatal write error.
func (s *Server) Serve(r io.Reader, w io.Writer) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 2*1024*1024)

	out := json.NewEncoder(w)

	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request

		err := json.Unmarshal(line, &req)
		if err != nil {
			slog.Warn("mcp: bad frame", "error", err)
			s.respond(out, nil, nil, &RPCError{Code: CodeParseError, Message: "not JSON"})

			continue
		}

		if req.ID == nil {
			continue // notification (e.g. initialized): silence
		}

		result, rpcErr := s.dispatch(req)

		s.respond(out, req.ID, result, rpcErr)
	}

	err := sc.Err()
	if err != nil {
		slog.Warn("mcp: stdin error", "error", err)
	}
}

func (s *Server) respond(out *json.Encoder, id any, result any, rpcErr *RPCError) {
	resp := Response{JSONRPC: "2.0", ID: id, Result: result, Error: rpcErr}

	err := out.Encode(resp)
	if err != nil {
		slog.Warn("mcp: stdout write failed", "error", err)
	}
}

// dispatch routes one method; tool failures come back as isError
// results (MCP style), protocol failures as RPC errors.
func (s *Server) dispatch(req Request) (any, *RPCError) {
	switch req.Method {
	case MethodInitialize:
		return s.hello(req.Params), nil

	case MethodToolsList:
		list := make([]Tool, 0, len(s.tools))
		for _, t := range s.tools {
			list = append(list, t)
		}

		return map[string]any{"tools": list}, nil

	case MethodToolsCall:
		return s.call(req.Params)

	case MethodPing:
		return map[string]any{}, nil

	default:
		return nil, &RPCError{Code: CodeMethodNotFound, Message: "no such method: " + req.Method}
	}
}

// hello answers initialize, echoing the client's protocol version
// (hosts reject servers that won't speak theirs).
func (s *Server) hello(params json.RawMessage) map[string]any {
	version := DefaultProtocol

	var p struct {
		ProtocolVersion string `json:"protocolVersion"` //nolint:tagliatelle // MCP wire names are camelCase by spec
	}

	err := json.Unmarshal(params, &p)
	if err == nil && p.ProtocolVersion != "" {
		version = p.ProtocolVersion
	}

	return map[string]any{
		"protocolVersion": version,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      s.info,
	}
}

// callParams is the tools/call envelope.
type callParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// call runs one tool; the text (or the failure) goes home inside
// a content block, never as a protocol error.
func (s *Server) call(params json.RawMessage) (any, *RPCError) {
	var p callParams

	err := json.Unmarshal(params, &p)
	if err != nil {
		return nil, &RPCError{Code: CodeInvalidParams, Message: "bad tools/call params"}
	}

	h, ok := s.run[p.Name]
	if !ok {
		return nil, &RPCError{Code: CodeMethodNotFound, Message: "no such tool: " + p.Name}
	}

	text, err := h(p.Arguments)
	if err != nil {
		return map[string]any{
			"content": []any{map[string]any{"type": "text", "text": err.Error()}},
			"isError": true,
		}, nil
	}

	return map[string]any{
		"content": []any{map[string]any{"type": "text", "text": text}},
	}, nil
}

// StrArg reads an optional string arg.
func StrArg(args map[string]any, key, def string) string {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return def
	}

	return v
}

// IntArg reads an optional numeric arg (JSON numbers arrive float64),
// clamped into [min, max] so callers can't fork-bomb the bench.
func IntArg(args map[string]any, key string, def, min, max int) int {
	v, ok := args[key].(float64)
	if !ok {
		return def
	}

	n := int(v)
	if n < min {
		return min
	}

	if n > max {
		return max
	}

	return n
}
