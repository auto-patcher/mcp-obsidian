package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func vaultTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_list_notes",
				mcp.WithDescription("List files and directories in the vault. Omit dirpath (or pass empty) to list the vault root."),
				mcp.WithString("dirpath",
					mcp.Description("Optional vault-relative directory. Empty or omitted lists the vault root. Empty directories are not returned."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				dirpath := argString(args, "dirpath")
				var (
					files []string
					err   error
				)
				if dirpath == "" {
					files, err = client.ListFilesInVault()
				} else {
					files, err = client.ListFilesInDir(dirpath)
				}
				if err != nil {
					return fail("list notes: %w", err)
				}
				return jsonOK(files)
			}),
		},
		{
			mcp.NewTool("obsidian_get_note",
				mcp.WithDescription(`Read one or more notes. Accepts a single filepath string or an array of filepaths.

By default returns the raw markdown content. When frontmatter_only is true, returns just the parsed YAML frontmatter as JSON ({} when the note has none). For batch reads (array input), the result is a JSON object keyed by filepath.`),
				// filepaths is intentionally untyped (accepts string or array); the
				// mcp-go builders don't expose a union type, so we validate at runtime.
				mcp.WithString("filepaths",
					mcp.Required(),
					mcp.Description("A single vault-relative filepath, or a JSON array of filepaths for batch reads."),
				),
				mcp.WithBoolean("frontmatter_only",
					mcp.Description("If true, return only the YAML frontmatter as JSON. Default: false."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				paths, err := argFilepaths(args, "filepaths")
				if err != nil {
					return fail("%w", err)
				}
				frontmatterOnly := argBool(args, "frontmatter_only", false)
				return getNoteResult(client, paths, frontmatterOnly)
			}),
		},
		{
			mcp.NewTool("obsidian_delete_note",
				mcp.WithDescription("Delete a file or directory from the vault. Errors if the path does not exist."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the file to delete (relative to vault root)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				if err := client.DeleteFile(fp); err != nil {
					return fail("delete note: %w", err)
				}
				return textOK("Successfully deleted " + fp)
			}),
		},
	}
}

// getNoteResult dispatches based on single vs batch and frontmatter-only mode.
func getNoteResult(client *obsidian.Client, paths []string, frontmatterOnly bool) (*mcp.CallToolResult, error) {
	if len(paths) == 1 {
		fp := paths[0]
		if frontmatterOnly {
			raw, err := client.GetFrontmatter(fp)
			if err != nil {
				return fail("get frontmatter: %w", err)
			}
			return rawOK(raw)
		}
		content, err := client.GetFileContents(fp)
		if err != nil {
			return fail("get note: %w", err)
		}
		return textOK(content)
	}
	if frontmatterOnly {
		result := make(map[string]json.RawMessage, len(paths))
		for _, fp := range paths {
			raw, err := client.GetFrontmatter(fp)
			if err != nil {
				result[fp] = json.RawMessage(`{"error":` + jsonString(err.Error()) + `}`)
				continue
			}
			result[fp] = raw
		}
		return jsonOK(result)
	}
	return textOK(client.GetBatchFileContents(paths))
}
