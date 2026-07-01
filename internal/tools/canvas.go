package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func canvasTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		// --- Canvas file operations -------------------------------------------
		{
			mcp.NewTool("obsidian_list_canvas_files",
				mcp.WithDescription("List all canvas (.canvas) files in the vault."),
			),
			noCtx(func(_ map[string]any) (*mcp.CallToolResult, error) {
				raw, err := client.ListCanvasFiles("")
				if err != nil {
					return fail("list canvas files: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_get_canvas",
				mcp.WithDescription("Read a canvas file and return its full JSON (nodes and edges arrays)."),
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
			mcp.NewTool("obsidian_write_canvas",
				mcp.WithDescription(`Create or completely overwrite a canvas file with the provided JSON data (nodes and edges arrays).`),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithObject("canvas",
					mcp.Required(),
					mcp.Description(`Canvas JSON with 'nodes' and 'edges' arrays. Example: {"nodes": [], "edges": []}`),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				canvasData, hasCanvas := args["canvas"]
				if !hasCanvas {
					return fail("canvas is required")
				}
				raw, err := client.WriteCanvas(fp, canvasData)
				if err != nil {
					return fail("write canvas: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_delete_canvas",
				mcp.WithDescription("Delete a canvas file from the vault."),
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
				if err := client.DeleteCanvas(fp); err != nil {
					return fail("delete canvas: %w", err)
				}
				return textOK("Successfully deleted canvas " + fp)
			}),
		},
		{
			mcp.NewTool("obsidian_search_canvases",
				mcp.WithDescription("Search canvas files by text query."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query string."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				query := argString(args, "query")
				if query == "" {
					return fail("query is required")
				}
				raw, err := client.SearchCanvases(query, "")
				if err != nil {
					return fail("search canvases: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_get_canvas_stats",
				mcp.WithDescription("Get statistics about a canvas file: node and edge counts by type, and the bounding box of all content."),
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
				raw, err := client.GetCanvasStats(fp)
				if err != nil {
					return fail("get canvas stats: %w", err)
				}
				return rawOK(raw)
			}),
		},

		// --- Node operations --------------------------------------------------
		{
			mcp.NewTool("obsidian_list_canvas_nodes",
				mcp.WithDescription("List all nodes in a canvas, optionally filtered by type (text, file, link, group)."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("type",
					mcp.Description("Filter by node type."),
					mcp.Enum("text", "file", "link", "group"),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				raw, err := client.ListCanvasNodes(fp, argString(args, "type"))
				if err != nil {
					return fail("list canvas nodes: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_get_canvas_node",
				mcp.WithDescription("Get a specific node from a canvas by its ID."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("node_id",
					mcp.Required(),
					mcp.Description("The ID of the node."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, nodeID := argString(args, "filepath"), argString(args, "node_id")
				if fp == "" || nodeID == "" {
					return fail("filepath and node_id are required")
				}
				raw, err := client.GetCanvasNode(fp, nodeID)
				if err != nil {
					return fail("get canvas node: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_create_canvas_node",
				mcp.WithDescription(`Create a new node in a canvas. Returns the node with its generated ID.

Required node fields: type (text|file|link|group), x, y, width, height.
Type-specific fields:
  text  → text (string)
  file  → file (vault path), subpath (optional heading/block)
  link  → url (string)
  group → label (optional), background, backgroundStyle`),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithObject("node",
					mcp.Required(),
					mcp.Description("Node data. Required: type, x, y, width, height."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				node, hasNode := args["node"]
				if fp == "" || !hasNode {
					return fail("filepath and node are required")
				}
				raw, err := client.CreateCanvasNode(fp, node)
				if err != nil {
					return fail("create canvas node: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_update_canvas_node",
				mcp.WithDescription("Update fields on an existing canvas node. Only provided fields are changed."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("node_id",
					mcp.Required(),
					mcp.Description("The ID of the node to update."),
				),
				mcp.WithObject("node",
					mcp.Required(),
					mcp.Description(`Fields to update on the node (e.g. {"x": 100, "text": "new content"}).`),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, nodeID := argString(args, "filepath"), argString(args, "node_id")
				node, hasNode := args["node"]
				if fp == "" || nodeID == "" || !hasNode {
					return fail("filepath, node_id, and node are required")
				}
				raw, err := client.UpdateCanvasNode(fp, nodeID, node)
				if err != nil {
					return fail("update canvas node: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_delete_canvas_node",
				mcp.WithDescription("Delete a node from a canvas. Optionally also deletes all edges connected to it."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("node_id",
					mcp.Required(),
					mcp.Description("The ID of the node to delete."),
				),
				mcp.WithBoolean("delete_edges",
					mcp.Description("Also delete edges connected to this node (default: false)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, nodeID := argString(args, "filepath"), argString(args, "node_id")
				if fp == "" || nodeID == "" {
					return fail("filepath and node_id are required")
				}
				if err := client.DeleteCanvasNode(fp, nodeID, argBool(args, "delete_edges", false)); err != nil {
					return fail("delete canvas node: %w", err)
				}
				return textOK("Successfully deleted node " + nodeID + " from " + fp)
			}),
		},

		// --- Edge operations --------------------------------------------------
		{
			mcp.NewTool("obsidian_list_canvas_edges",
				mcp.WithDescription("List all edges in a canvas."),
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
				raw, err := client.ListCanvasEdges(fp)
				if err != nil {
					return fail("list canvas edges: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_get_canvas_edge",
				mcp.WithDescription("Get a specific edge from a canvas by its ID."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("edge_id",
					mcp.Required(),
					mcp.Description("The ID of the edge."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, edgeID := argString(args, "filepath"), argString(args, "edge_id")
				if fp == "" || edgeID == "" {
					return fail("filepath and edge_id are required")
				}
				raw, err := client.GetCanvasEdge(fp, edgeID)
				if err != nil {
					return fail("get canvas edge: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_create_canvas_edge",
				mcp.WithDescription(`Create a new edge between two nodes. Returns the edge with its generated ID.

Required: fromNode, toNode (node IDs).
Optional: fromSide/toSide (top|right|bottom|left), fromEnd/toEnd (none|arrow), color, label.`),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithObject("edge",
					mcp.Required(),
					mcp.Description("Edge data. Required: fromNode, toNode."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				edge, hasEdge := args["edge"]
				if fp == "" || !hasEdge {
					return fail("filepath and edge are required")
				}
				raw, err := client.CreateCanvasEdge(fp, edge)
				if err != nil {
					return fail("create canvas edge: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_update_canvas_edge",
				mcp.WithDescription("Update fields on an existing canvas edge."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("edge_id",
					mcp.Required(),
					mcp.Description("The ID of the edge to update."),
				),
				mcp.WithObject("edge",
					mcp.Required(),
					mcp.Description("Fields to update on the edge."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, edgeID := argString(args, "filepath"), argString(args, "edge_id")
				edge, hasEdge := args["edge"]
				if fp == "" || edgeID == "" || !hasEdge {
					return fail("filepath, edge_id, and edge are required")
				}
				raw, err := client.UpdateCanvasEdge(fp, edgeID, edge)
				if err != nil {
					return fail("update canvas edge: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_delete_canvas_edge",
				mcp.WithDescription("Delete an edge from a canvas."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("edge_id",
					mcp.Required(),
					mcp.Description("The ID of the edge to delete."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, edgeID := argString(args, "filepath"), argString(args, "edge_id")
				if fp == "" || edgeID == "" {
					return fail("filepath and edge_id are required")
				}
				if err := client.DeleteCanvasEdge(fp, edgeID); err != nil {
					return fail("delete canvas edge: %w", err)
				}
				return textOK("Successfully deleted edge " + edgeID + " from " + fp)
			}),
		},

		// --- Group operations -------------------------------------------------
		{
			mcp.NewTool("obsidian_list_canvas_groups",
				mcp.WithDescription("List all group nodes in a canvas."),
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
				raw, err := client.ListCanvasGroups(fp)
				if err != nil {
					return fail("list canvas groups: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_get_canvas_group",
				mcp.WithDescription("Get a specific group node and the nodes it contains spatially."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("group_id",
					mcp.Required(),
					mcp.Description("The ID of the group."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, groupID := argString(args, "filepath"), argString(args, "group_id")
				if fp == "" || groupID == "" {
					return fail("filepath and group_id are required")
				}
				raw, err := client.GetCanvasGroup(fp, groupID)
				if err != nil {
					return fail("get canvas group: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_create_canvas_group",
				mcp.WithDescription(`Create a new group node in a canvas. Returns the group with its generated ID.

Required fields: type (must be "group"), x, y, width, height.
Optional: label, color, background, backgroundStyle.`),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithObject("group",
					mcp.Required(),
					mcp.Description(`Group node data. Required: type="group", x, y, width, height.`),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				group, hasGroup := args["group"]
				if fp == "" || !hasGroup {
					return fail("filepath and group are required")
				}
				raw, err := client.CreateCanvasGroup(fp, group)
				if err != nil {
					return fail("create canvas group: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_update_canvas_group",
				mcp.WithDescription("Update fields on an existing canvas group node."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("group_id",
					mcp.Required(),
					mcp.Description("The ID of the group to update."),
				),
				mcp.WithObject("group",
					mcp.Required(),
					mcp.Description("Fields to update on the group."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, groupID := argString(args, "filepath"), argString(args, "group_id")
				group, hasGroup := args["group"]
				if fp == "" || groupID == "" || !hasGroup {
					return fail("filepath, group_id, and group are required")
				}
				raw, err := client.UpdateCanvasGroup(fp, groupID, group)
				if err != nil {
					return fail("update canvas group: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_delete_canvas_group",
				mcp.WithDescription("Delete a group node from a canvas."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the .canvas file (relative to vault root)."),
				),
				mcp.WithString("group_id",
					mcp.Required(),
					mcp.Description("The ID of the group to delete."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp, groupID := argString(args, "filepath"), argString(args, "group_id")
				if fp == "" || groupID == "" {
					return fail("filepath and group_id are required")
				}
				if err := client.DeleteCanvasGroup(fp, groupID); err != nil {
					return fail("delete canvas group: %w", err)
				}
				return textOK("Successfully deleted group " + groupID + " from " + fp)
			}),
		},
	}
}
