package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func searchTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_search_notes",
				mcp.WithDescription(`Search vault notes. Three modes:

- 'text' (default): simple full-text search. Uses 'query' (string) and optional 'context_length'.
- 'tag': find notes carrying a specific tag. Use 'query' for the tag name (no leading '#'); hierarchical matching is exact ('work' does not match 'work/tasks'). Optional 'dirpath' scopes results.
- 'jsonlogic': advanced search via a JsonLogic query in the 'jsonlogic' field. Supports standard operators plus 'glob' and 'regexp'.

JsonLogic examples:
  All markdown files:
    {"glob": ["*.md", {"var": "path"}]}
  Markdown files containing "Keaton" in Work/:
    {"and": [{"glob": ["*.md", {"var": "path"}]}, {"regexp": [".*Work.*", {"var": "path"}]}, {"regexp": ["Keaton", {"var": "content"}]}]}`),
				mcp.WithString("query",
					mcp.Description("Text query (mode='text') or tag name without '#' (mode='tag'). Ignored when mode='jsonlogic'."),
				),
				mcp.WithString("mode",
					mcp.Description("Search mode. Default: text."),
					mcp.Enum("text", "tag", "jsonlogic"),
				),
				mcp.WithObject("jsonlogic",
					mcp.Description("JsonLogic query object. Required when mode='jsonlogic'."),
				),
				mcp.WithNumber("context_length",
					mcp.Description("Amount of context around each match (mode='text' only, default 100)."),
				),
				mcp.WithString("dirpath",
					mcp.Description("Optional vault-relative directory to scope results (mode='tag' only)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				mode := argString(args, "mode")
				if mode == "" {
					mode = "text"
				}
				switch mode {
				case "text":
					query := argString(args, "query")
					if query == "" {
						return fail("query is required for mode='text'")
					}
					raw, err := client.Search(query, argInt(args, "context_length", 100))
					if err != nil {
						return fail("search: %w", err)
					}
					return rawOK(raw)
				case "tag":
					tag := argString(args, "query")
					if tag == "" {
						return fail("query (tag name) is required for mode='tag'")
					}
					paths, err := client.SearchByTag(tag, argString(args, "dirpath"))
					if err != nil {
						return fail("search by tag: %w", err)
					}
					return jsonOK(paths)
				case "jsonlogic":
					jl, hasJL := args["jsonlogic"]
					if !hasJL {
						return fail("jsonlogic is required for mode='jsonlogic'")
					}
					raw, err := client.SearchJSON(jl)
					if err != nil {
						return fail("jsonlogic search: %w", err)
					}
					return rawOK(raw)
				default:
					return fail("unknown mode %q (expected text|tag|jsonlogic)", mode)
				}
			}),
		},
	}
}
