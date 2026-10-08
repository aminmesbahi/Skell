// Package mcp implements a minimal Model Context Protocol server over stdio
// (newline-delimited JSON-RPC 2.0). It has no dependencies beyond the
// standard library and supports what tool-only servers need: initialize,
// ping, tools/list and tools/call.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sync"
)

// SupportedVersions lists MCP protocol revisions this server speaks, newest
// first. The client's requested version is echoed when supported.
var SupportedVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// Tool is one callable tool.
type Tool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	Annotations *Annotations   `json:"annotations,omitempty"`
	// Handler runs the tool. A returned error becomes a tool result with
	// isError=true (so the model can see and react to it).
	Handler func(ctx context.Context, args json.RawMessage) (any, error) `json:"-"`
}

// Annotations are hints clients use to decide how much to confirm.
type Annotations struct {
	ReadOnlyHint    bool `json:"readOnlyHint,omitempty"`
	DestructiveHint bool `json:"destructiveHint,omitempty"`
	IdempotentHint  bool `json:"idempotentHint,omitempty"`
	OpenWorldHint   bool `json:"openWorldHint,omitempty"`
}

// Server is a stdio MCP server.
type Server struct {
	Name         string
	Version      string
	Instructions string
	tools        []Tool
}

// NewServer creates a server with the given tools.
func NewServer(name, version, instructions string, tools []Tool) *Server {
	return &Server{Name: name, Version: version, Instructions: instructions, tools: tools}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	errParse          = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
)

// Serve reads requests from r and writes responses to w until r is closed or
// ctx is cancelled. Requests are handled one at a time, in order.
func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	var mu sync.Mutex
	enc := json.NewEncoder(w)
	write := func(v any) {
		mu.Lock()
		defer mu.Unlock()
		_ = enc.Encode(v)
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 16<<20)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			write(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{Code: errParse, Message: "parse error"}})
			continue
		}
		if req.JSONRPC != "2.0" || req.Method == "" {
			if len(req.ID) > 0 {
				write(response{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: errInvalidRequest, Message: "invalid request"}})
			}
			continue
		}
		isNotification := len(req.ID) == 0
		result, rerr := s.handle(ctx, req)
		if isNotification {
			continue
		}
		resp := response{JSONRPC: "2.0", ID: req.ID}
		if rerr != nil {
			resp.Error = rerr
		} else {
			resp.Result = result
		}
		write(resp)
	}
	return sc.Err()
}

func (s *Server) handle(ctx context.Context, req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		version := SupportedVersions[0]
		if slices.Contains(SupportedVersions, p.ProtocolVersion) {
			version = p.ProtocolVersion
		}
		return map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": s.Name, "version": s.Version},
			"instructions":    s.Instructions,
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": s.tools}, nil
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, &rpcError{Code: errInvalidParams, Message: "invalid params"}
		}
		for _, t := range s.tools {
			if t.Name != p.Name {
				continue
			}
			args := p.Arguments
			if len(args) == 0 || string(args) == "null" {
				args = json.RawMessage("{}")
			}
			out, err := runTool(ctx, t, args)
			if err != nil {
				return toolResult(err.Error(), true), nil
			}
			return toolResult(out, false), nil
		}
		return nil, &rpcError{Code: errInvalidParams, Message: fmt.Sprintf("unknown tool %q", p.Name)}
	default:
		if len(req.Method) > 14 && req.Method[:14] == "notifications/" {
			return nil, nil
		}
		return nil, &rpcError{Code: errMethodNotFound, Message: "method not found: " + req.Method}
	}
}

// runTool calls a handler, converting panics into errors so one bad call
// can't take the server down.
func runTool(ctx context.Context, t Tool, args json.RawMessage) (out any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("internal error in %s: %v", t.Name, r)
		}
	}()
	return t.Handler(ctx, args)
}

func toolResult(v any, isError bool) map[string]any {
	text, ok := v.(string)
	if !ok {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			text = fmt.Sprint(v)
		} else {
			text = string(b)
		}
	}
	res := map[string]any{"content": []map[string]any{{"type": "text", "text": text}}}
	if isError {
		res["isError"] = true
	}
	return res
}

// Object builds a JSON-schema object with the given properties and required keys.
func Object(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

// Str is a JSON-schema string property.
func Str(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }

// StrList is a JSON-schema array-of-strings property.
func StrList(desc string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
}

// Bool is a JSON-schema boolean property.
func Bool(desc string) map[string]any { return map[string]any{"type": "boolean", "description": desc} }
