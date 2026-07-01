package tools

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
)

func periodicTools(client *obsidian.Client) []toolEntry {
	return []toolEntry{
		{
			mcp.NewTool("obsidian_get_periodic_note",
				mcp.WithDescription("Get the current periodic note for the given period, or the N most recent periodic notes when 'recent' is set."),
				mcp.WithString("period",
					mcp.Required(),
					mcp.Description("Period type."),
					mcp.Enum("daily", "weekly", "monthly", "quarterly", "yearly"),
				),
				mcp.WithNumber("recent",
					mcp.Description("If set, return this many most-recent periodic notes instead of the current one."),
				),
				mcp.WithString("type",
					mcp.Description("'content' returns markdown; 'metadata' includes note metadata. Default: content. (Current-note mode only.)"),
					mcp.Enum("content", "metadata"),
				),
				mcp.WithBoolean("include_content",
					mcp.Description("When 'recent' is set, whether to include note content (default: false)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				period := argString(args, "period")
				if period == "" {
					return fail("period is required")
				}
				if _, hasRecent := args["recent"]; hasRecent {
					limit := argInt(args, "recent", 5)
					if limit < 1 {
						return fail("recent must be a positive integer")
					}
					raw, err := client.GetRecentPeriodicNotes(period, limit, argBool(args, "include_content", false))
					if err != nil {
						return fail("get recent periodic notes: %w", err)
					}
					return rawOK(raw)
				}
				noteType := argString(args, "type")
				if noteType == "" {
					noteType = "content"
				}
				content, err := client.GetPeriodicNote(period, noteType)
				if err != nil {
					return fail("get periodic note: %w", err)
				}
				return textOK(content)
			}),
		},
		{
			mcp.NewTool("obsidian_get_recent_changes",
				mcp.WithDescription("Get recently modified files in the vault."),
				mcp.WithNumber("limit",
					mcp.Description("Maximum number of files to return (default: 10)."),
				),
				mcp.WithNumber("days",
					mcp.Description("Only include files modified within this many days (default: 90)."),
				),
			),
			noCtx(func(args map[string]any) (*mcp.CallToolResult, error) {
				limit := argInt(args, "limit", 10)
				days := argInt(args, "days", 90)
				if limit < 1 {
					return fail("limit must be a positive integer")
				}
				if days < 1 {
					return fail("days must be a positive integer")
				}
				raw, err := client.GetRecentChanges(limit, days)
				if err != nil {
					return fail("get recent changes: %w", err)
				}
				return rawOK(raw)
			}),
		},
	}
}
