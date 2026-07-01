package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func contentTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_append_content",
				mcp.WithDescription("Append content to a new or existing file in the vault."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the file (relative to vault root)."),
				),
				mcp.WithString("content",
					mcp.Required(),
					mcp.Description("Content to append to the file."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				content := argString(args, "content")
				if fp == "" || content == "" {
					return fail("filepath and content are required")
				}
				if err := client.AppendContent(fp, content); err != nil {
					return fail("append content: %w", err)
				}
				return textOK("Successfully appended content to " + fp)
			}),
		},
		{
			mcp.NewTool("obsidian_patch_content",
				mcp.WithDescription(`Insert content into an existing note relative to a heading, block reference, or frontmatter field.

For target_type='heading', target is the fully-qualified heading path with '::' separator (e.g. 'Outer::Inner'). Bare heading names (no '::') are auto-qualified if they match exactly one heading in the file.`),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the file (relative to vault root)."),
				),
				mcp.WithString("operation",
					mcp.Required(),
					mcp.Description("Operation to perform."),
					mcp.Enum("append", "prepend", "replace"),
				),
				mcp.WithString("target_type",
					mcp.Required(),
					mcp.Description("Type of target to patch."),
					mcp.Enum("heading", "block", "frontmatter"),
				),
				mcp.WithString("target",
					mcp.Required(),
					mcp.Description("Target identifier. For 'heading': qualified path with '::' (e.g. 'Section::Subsection'), bare name auto-qualifies if unambiguous. For 'block': block reference id. For 'frontmatter': YAML field name."),
				),
				mcp.WithString("content",
					mcp.Required(),
					mcp.Description("Content to insert."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				operation := argString(args, "operation")
				targetType := argString(args, "target_type")
				target := argString(args, "target")
				content := argString(args, "content")
				if fp == "" || operation == "" || targetType == "" || target == "" {
					return fail("filepath, operation, target_type, and target are required")
				}
				if err := client.PatchContent(fp, operation, targetType, target, content); err != nil {
					return fail("patch content: %w", err)
				}
				return textOK("Successfully patched content in " + fp)
			}),
		},
		{
			mcp.NewTool("obsidian_put_content",
				mcp.WithDescription("Create a new file or COMPLETELY OVERWRITE an existing file. Previous content is lost. Use obsidian_append_content to add without erasing, or obsidian_patch_content to modify a specific section."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the file (relative to vault root)."),
				),
				mcp.WithString("content",
					mcp.Required(),
					mcp.Description("Full file content. Replaces all existing content."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				content := argString(args, "content")
				if fp == "" || content == "" {
					return fail("filepath and content are required")
				}
				if err := client.PutContent(fp, content); err != nil {
					return fail("put content: %w", err)
				}
				return textOK("Successfully wrote content to " + fp)
			}),
		},
	}
}
