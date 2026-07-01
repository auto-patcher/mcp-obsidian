package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/auto-patcher/mcp-obsidian/internal/obsidian"
	"github.com/auto-patcher/mcp-obsidian/internal/tools"
)

func main() {
	cfg := obsidian.ConfigFromEnv()
	isHTTPS := strings.HasPrefix(cfg.BaseURL, "https://") || cfg.Protocol == "https"
	if isHTTPS && cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "warning: OBSIDIAN_API_KEY not set but HTTPS is configured \u2014 requests will be unauthenticated")
	}

	client := obsidian.New(cfg)
	s := server.NewMCPServer("mcp-obsidian", "0.3.0")
	tools.RegisterAll(s, client)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
