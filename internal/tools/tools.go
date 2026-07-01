package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

// handlerFn is the MCP tool handler signature.
type handlerFn = server.ToolHandlerFunc

type toolEntry struct {
	tool    mcp.Tool
	handler handlerFn
}

// noCtx wraps a simple args-only function as a full MCP ToolHandlerFunc.
// For handlers that need the context (cancellation, deadlines), use handlerFn directly.
func noCtx(f func(args map[string]any) (*mcp.CallToolResult, error)) handlerFn {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return f(req.Params.Arguments)
	}
}

// RegisterAll registers every Obsidian tool with the MCP server.
// To add a new group of tools, create a new *Tools(client) function and
// append it to the slice in allTools.
func RegisterAll(s *server.MCPServer, client *obsidian.Client) {
	for _, e := range allTools(client) {
		s.AddTool(e.tool, e.handler)
	}
}

func allTools(client *obsidian.Client) []toolEntry {
	var all []toolEntry
	for _, g := range [][]toolEntry{
		vaultTools(client),
		searchTools(client),
		contentTools(client),
		periodicTools(client),
		canvasTools(client),
	} {
		all = append(all, g...)
	}
	return all
}

// --- argument helpers --------------------------------------------------------

func argString(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return v
}

func argInt(args map[string]any, key string, def int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return def
	}
}

func argBool(args map[string]any, key string, def bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return def
}

// argFilepaths accepts a single string or an array of strings and returns a
// non-empty slice. Returns an error if the field is missing or empty.
func argFilepaths(args map[string]any, key string) ([]string, error) {
	v, ok := args[key]
	if !ok {
		return nil, fmt.Errorf("%s is required", key)
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil, fmt.Errorf("%s is required", key)
		}
		return []string{t}, nil
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("%s must be a non-empty string or array of strings", key)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s must be a string or array of strings", key)
	}
}

// jsonString encodes s as a JSON string literal (including quotes and escapes).
func jsonString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

// --- result helpers ----------------------------------------------------------

// textOK returns a plain-text tool result.
func textOK(s string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(s), nil
}

// jsonOK marshals v as indented JSON and returns a text result.
func jsonOK(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fail("marshal result: %w", err)
	}
	return textOK(string(data))
}

// rawOK returns a pre-marshaled JSON payload as a text result.
func rawOK(raw json.RawMessage) (*mcp.CallToolResult, error) {
	return textOK(string(raw))
}

// fail returns a formatted error (the MCP server surfaces it to the client).
func fail(format string, a ...any) (*mcp.CallToolResult, error) {
	return nil, fmt.Errorf(format, a...)
}
