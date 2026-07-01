package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func canvasTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_get_canvas",
				mcp.WithDescription("Read a canvas (.canvas) file and return its full JSON: 'nodes' (text/file/link/group cards with spatial layout) and 'edges' (connections between nodes)."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				raw, err := client.GetCanvas(fp)
				if err != nil {
					return fail("get canvas: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_patch_canvas",
				mcp.WithDescription(`Modify a canvas by applying an ordered list of operations. Creates the canvas if it does not exist.

Each operation is an object with an 'op' field. Operations execute sequentially; the response reports per-op status and any auto-generated IDs.

Operation shapes:
  { "op": "add_node",    "type": "text|file|link|group", ...spatial fields (x, y, width, height) and type-specific fields }
    text  -> "text": string
    file  -> "file": vault path, optional "subpath"
    link  -> "url": string
    group -> optional "label", "background", "backgroundStyle"
  { "op": "update_node", "id": "...", ...fields to change }
  { "op": "remove_node", "id": "..." }   // also removes connected edges
  { "op": "add_edge",    "fromNode": "...", "toNode": "...", optional "fromSide"/"toSide" (top|right|bottom|left), "fromEnd"/"toEnd" (none|arrow), "color", "label" }
  { "op": "update_edge", "id": "...", ...fields to change }
  { "op": "remove_edge", "id": "..." }

If an operation fails mid-list, earlier operations are NOT rolled back — the response indicates which op failed.`),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root). Created if it does not exist."),
				),
				mcp.WithArray("operations",
					mcp.Required(),
					mcp.Description("Ordered list of operations to apply."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				opsRaw, ok := args["operations"].([]any)
				if !ok || len(opsRaw) == 0 {
					return fail("operations must be a non-empty array")
				}
				if err := ensureCanvasExists(client, fp); err != nil {
					return fail("ensure canvas: %w", err)
				}
				return applyCanvasOps(client, fp, opsRaw)
			}),
		},
	}
}

// ensureCanvasExists creates an empty canvas file if the given path does not
// yet resolve. Any GET error other than 404 is returned.
func ensureCanvasExists(client *obsidian.Client, fp string) error {
	_, err := client.GetCanvas(fp)
	if err == nil {
		return nil
	}
	if !isNotFound(err) {
		return err
	}
	_, werr := client.WriteCanvas(fp, map[string]any{
		"nodes": []any{},
		"edges": []any{},
	})
	return werr
}

// isNotFound heuristically detects Obsidian REST 404 responses. The client
// surfaces API errors as "Error <code>: <message>" or "HTTP 404: ..." strings.
func isNotFound(err error) bool {
	s := err.Error()
	return strings.Contains(s, "HTTP 404") || strings.Contains(s, "Error 40400") || strings.Contains(s, "does not exist")
}

// applyCanvasOps dispatches each operation to the corresponding client method
// and collects per-op results. Stops on the first error.
func applyCanvasOps(client *obsidian.Client, fp string, ops []any) (*mcp.CallToolResult, error) {
	results := make([]map[string]any, 0, len(ops))
	for i, raw := range ops {
		op, ok := raw.(map[string]any)
		if !ok {
			return fail("operations[%d]: expected object", i)
		}
		kind := argString(op, "op")
		res, err := dispatchCanvasOp(client, fp, kind, op)
		entry := map[string]any{"index": i, "op": kind}
		if err != nil {
			entry["error"] = err.Error()
			results = append(results, entry)
			return jsonOK(map[string]any{
				"success":     false,
				"failed_at":   i,
				"results":     results,
				"error":       err.Error(),
			})
		}
		for k, v := range res {
			entry[k] = v
		}
		results = append(results, entry)
	}
	return jsonOK(map[string]any{
		"success": true,
		"results": results,
	})
}

// dispatchCanvasOp performs a single canvas operation and returns a map of
// result fields (e.g. generated ID) to merge into the response entry.
func dispatchCanvasOp(client *obsidian.Client, fp, kind string, op map[string]any) (map[string]any, error) {
	switch kind {
	case "add_node":
		node := copyWithoutOp(op)
		raw, err := client.CreateCanvasNode(fp, node)
		if err != nil {
			return nil, err
		}
		return map[string]any{"id": extractID(raw), "node": json.RawMessage(raw)}, nil
	case "update_node":
		id := argString(op, "id")
		if id == "" {
			return nil, fmt.Errorf("update_node requires 'id'")
		}
		updates := copyWithoutOpAndID(op)
		raw, err := client.UpdateCanvasNode(fp, id, updates)
		if err != nil {
			return nil, err
		}
		return map[string]any{"id": id, "node": json.RawMessage(raw)}, nil
	case "remove_node":
		id := argString(op, "id")
		if id == "" {
			return nil, fmt.Errorf("remove_node requires 'id'")
		}
		if err := client.DeleteCanvasNode(fp, id, true); err != nil {
			return nil, err
		}
		return map[string]any{"id": id}, nil
	case "add_edge":
		edge := copyWithoutOp(op)
		raw, err := client.CreateCanvasEdge(fp, edge)
		if err != nil {
			return nil, err
		}
		return map[string]any{"id": extractID(raw), "edge": json.RawMessage(raw)}, nil
	case "update_edge":
		id := argString(op, "id")
		if id == "" {
			return nil, fmt.Errorf("update_edge requires 'id'")
		}
		updates := copyWithoutOpAndID(op)
		raw, err := client.UpdateCanvasEdge(fp, id, updates)
		if err != nil {
			return nil, err
		}
		return map[string]any{"id": id, "edge": json.RawMessage(raw)}, nil
	case "remove_edge":
		id := argString(op, "id")
		if id == "" {
			return nil, fmt.Errorf("remove_edge requires 'id'")
		}
		if err := client.DeleteCanvasEdge(fp, id); err != nil {
			return nil, err
		}
		return map[string]any{"id": id}, nil
	default:
		return nil, fmt.Errorf("unknown op %q (expected add_node|update_node|remove_node|add_edge|update_edge|remove_edge)", kind)
	}
}

func copyWithoutOp(op map[string]any) map[string]any {
	out := make(map[string]any, len(op))
	for k, v := range op {
		if k == "op" {
			continue
		}
		out[k] = v
	}
	return out
}

func copyWithoutOpAndID(op map[string]any) map[string]any {
	out := make(map[string]any, len(op))
	for k, v := range op {
		if k == "op" || k == "id" {
			continue
		}
		out[k] = v
	}
	return out
}

// extractID pulls the "id" field out of a returned node/edge JSON. Returns
// empty string if not present or unparseable.
func extractID(raw json.RawMessage) string {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return payload.ID
}
