package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
	"github.com/auto-patcher/mcp-obsidian/internal/tools"
)

func main() {
	cfg := obsidian.ConfigFromEnv()
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "error: OBSIDIAN_API_KEY environment variable required")
		os.Exit(1)
	}

	client := obsidian.New(cfg)
	s := server.NewMCPServer("mcp-obsidian", "0.3.0")
	tools.RegisterAll(s, client)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
