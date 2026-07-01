package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
	"github.com/auto-patcher/mcp-obsidian/internal/tools"
)

func main() {
	cfg := obsidian.ConfigFromEnv()
	u, _ := url.Parse(cfg.BaseURL)
	if u.Scheme == "https" && cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "warning: OBSIDIAN_API_KEY not set but HTTPS is configured — requests will be unauthenticated")
	}

	client := obsidian.New(cfg)
	s := server.NewMCPServer("mcp-obsidian", "0.3.0")
	tools.RegisterAll(s, client)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
