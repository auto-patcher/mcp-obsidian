package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func searchTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_simple_search",
				mcp.WithDescription("Simple full-text search across all files in the vault."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Text to search for in the vault."),
				),
				mcp.WithNumber("context_length",
					mcp.Description("Amount of context to return around each match (default: 100)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				query := argString(args, "query")
				if query == "" {
					return fail("query is required")
				}
				raw, err := client.Search(query, argInt(args, "context_length", 100))
				if err != nil {
					return fail("search: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_complex_search",
				mcp.WithDescription(`Complex search using a JsonLogic query. Supports standard JsonLogic operators plus 'glob' and 'regexp' for pattern matching.

Examples:
  All markdown files:
    {"glob": ["*.md", {"var": "path"}]}

  Markdown files containing "1221":
    {"and": [{"glob": ["*.md", {"var": "path"}]}, {"regexp": [".*1221.*", {"var": "content"}]}]}

  Markdown files in Work/ containing "Keaton":
    {"and": [{"glob": ["*.md", {"var": "path"}]}, {"regexp": [".*Work.*", {"var": "path"}]}, {"regexp": ["Keaton", {"var": "content"}]}]}`),
				mcp.WithObject("query",
					mcp.Required(),
					mcp.Description("JsonLogic query object. See tool description for examples."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				query, hasQuery := args["query"]
				if !hasQuery {
					return fail("query is required")
				}
				raw, err := client.SearchJSON(query)
				if err != nil {
					return fail("complex search: %w", err)
				}
				return rawOK(raw)
			}),
		},
		{
			mcp.NewTool("obsidian_search_by_tag",
				mcp.WithDescription("Find all notes carrying a specific tag. Matches the note's parsed tag set (YAML frontmatter 'tags:' plus inline '#tag' occurrences). Pass the tag without '#'. Hierarchical matching is exact — 'work' does not match 'work/tasks'."),
				mcp.WithString("tag",
					mcp.Required(),
					mcp.Description("Tag name without '#' (e.g. 'project', 'work/tasks')."),
				),
				mcp.WithString("dirpath",
					mcp.Description("Optional vault-relative directory to scope results (e.g. 'work/projects')."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				tag := argString(args, "tag")
				if tag == "" {
					return fail("tag is required")
				}
				paths, err := client.SearchByTag(tag, argString(args, "dirpath"))
				if err != nil {
					return fail("search by tag: %w", err)
				}
				return jsonOK(paths)
			}),
		},
	}
}
