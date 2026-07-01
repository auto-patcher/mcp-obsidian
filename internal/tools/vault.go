package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func vaultTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_list_files_in_vault",
				mcp.WithDescription("Lists all files and directories in the root directory of your Obsidian vault."),
			),
			noCtx(func(_ map[string]any) (*mcp.CallToolResult, error) {
				files, err := client.ListFilesInVault()
				if err != nil {
					return fail("list vault: %w", err)
				}
				return jsonOK(files)
			}),
		},
		{
			mcp.NewTool("obsidian_list_files_in_dir",
				mcp.WithDescription("Lists all files and directories in a specific Obsidian vault directory."),
				mcp.WithString("dirpath",
					mcp.Required(),
					mcp.Description("Path to list files from (relative to vault root). Empty directories are not returned."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				dirpath := argString(args, "dirpath")
				if dirpath == "" {
					return fail("dirpath is required")
				}
				files, err := client.ListFilesInDir(dirpath)
				if err != nil {
					return fail("list dir: %w", err)
				}
				return jsonOK(files)
			}),
		},
		{
			mcp.NewTool("obsidian_get_file_contents",
				mcp.WithDescription("Return the content of a single file in your vault."),
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
				content, err := client.GetFileContents(fp)
				if err != nil {
					return fail("get file: %w", err)
				}
				return textOK(content)
			}),
		},
		{
			mcp.NewTool("obsidian_batch_get_file_contents",
				mcp.WithDescription("Return the contents of multiple files concatenated with markdown headers."),
				mcp.WithArray("filepaths",
					mcp.Required(),
					mcp.Description("List of file paths to read (relative to vault root)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fps := argStringSlice(args, "filepaths")
				if len(fps) == 0 {
					return fail("filepaths must be a non-empty array")
				}
				return textOK(client.GetBatchFileContents(fps))
			}),
		},
		{
			mcp.NewTool("obsidian_delete_file",
				mcp.WithDescription("Delete a file or directory from the vault."),
				mcp.WithString("filepath",
					mcp.Required(),
					mcp.Description("Path to the file to delete (relative to vault root)."),
				),
				mcp.WithBoolean("confirm",
					mcp.Required(),
					mcp.Description("Must be true to confirm deletion."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				fp := argString(args, "filepath")
				if fp == "" {
					return fail("filepath is required")
				}
				if !argBool(args, "confirm", false) {
					return fail("confirm must be true to delete a file")
				}
				if err := client.DeleteFile(fp); err != nil {
					return fail("delete file: %w", err)
				}
				return textOK("Successfully deleted " + fp)
			}),
		},
	}
}
