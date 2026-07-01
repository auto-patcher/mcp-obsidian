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
		frontmatterTools(client),
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

func argStringSlice(args map[string]any, key string) []string {
	arr, ok := args[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
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
