package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func run(t *testing.T, s *Server, lines ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, s.Serve(context.Background(), strings.NewReader(strings.Join(lines, "\n")+"\n"), &out))
	var resps []map[string]any
	dec := json.NewDecoder(&out)
	for dec.More() {
		var m map[string]any
		require.NoError(t, dec.Decode(&m))
		resps = append(resps, m)
	}
	return resps
}

func testServer() *Server {
	return NewServer("skell", "test", "hi", []Tool{
		{
			Name: "echo", Description: "echo", InputSchema: Object(map[string]any{"x": Str("x")}, "x"),
			Handler: func(_ context.Context, args json.RawMessage) (any, error) {
				var a struct{ X string }
				_ = json.Unmarshal(args, &a)
				return map[string]string{"x": a.X}, nil
			},
		},
		{
			Name: "fail", Description: "fails", InputSchema: Object(map[string]any{}),
			Handler: func(context.Context, json.RawMessage) (any, error) { return nil, errors.New("boom") },
		},
		{
			Name: "panic", Description: "panics", InputSchema: Object(map[string]any{}),
			Handler: func(context.Context, json.RawMessage) (any, error) { panic("oops") },
		},
	})
}

func TestServer_Lifecycle(t *testing.T) {
	resps := run(t, testServer(),
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"x":"hello"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"fail"}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"panic"}}`,
		`{"jsonrpc":"2.0","id":6,"method":"nope"}`,
		`not json`,
		`{"jsonrpc":"2.0","id":7,"method":"ping"}`,
	)
	require.Len(t, resps, 8, "notification gets no response")

	init := resps[0]["result"].(map[string]any)
	assert.Equal(t, "2025-03-26", init["protocolVersion"])

	tools := resps[1]["result"].(map[string]any)["tools"].([]any)
	assert.Len(t, tools, 3)

	call := resps[2]["result"].(map[string]any)
	text := call["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, text, "hello")
	assert.Nil(t, call["isError"])

	assert.Equal(t, true, resps[3]["result"].(map[string]any)["isError"])
	assert.Equal(t, true, resps[4]["result"].(map[string]any)["isError"])
	assert.Equal(t, float64(errMethodNotFound), resps[5]["error"].(map[string]any)["code"])
	assert.Equal(t, float64(errParse), resps[6]["error"].(map[string]any)["code"])
	assert.NotNil(t, resps[7]["result"])
}

func TestServer_UnknownProtocolVersionGetsLatest(t *testing.T) {
	resps := run(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1999-01-01"}}`)
	assert.Equal(t, SupportedVersions[0], resps[0]["result"].(map[string]any)["protocolVersion"])
}
