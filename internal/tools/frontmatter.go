package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func frontmatterTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_get_frontmatter",
				mcp.WithDescription("Return the YAML frontmatter of a note as a parsed JSON object. Lighter than obsidian_get_file_contents when you only need metadata. Returns {} for notes without frontmatter."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the file (relative to vault root)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				raw, err := client.GetFrontmatter(fp)
				if err != nil {
					return fail("get frontmatter: %w", err)
				}
				return rawOK(raw)
			}),
		},
	}
}
