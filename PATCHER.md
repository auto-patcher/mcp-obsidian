# PATCHER.md

This file helps the autopatcher understand how this fork diverges from the upstream repository and how to maintain it.

## Repositories

**Upstream:** https://github.com/auto-patcher/mcp-obsidian
**Fork:** https://github.com/auto-patcher/mcp-obsidian (local copy at `/home/eva/sources/mcp-obsidian`)

## Upstream Baseline

**Last patched:** Latest commit (development version)

The fork is kept in sync with upstream development. The autopatcher can rebase on main without conflicts for most changes.

## Purpose

This MCP server bridges the gap between the Python/CLI ecosystem and Obsidian, allowing Python scripts, CLI tools, and any MCP-compatible client to interact with an Obsidian vault through the Obsidian Local REST API plugin.

Key use cases:
- Use Obsidian vault as a data source from Python scripts and CLI tools
- Integrate Obsidian with AI agents that prefer Python tooling
- Extend MCP capabilities for Obsidian in non-Node.js environments
- Enable bot/automation workflows over MCP protocol

## Character

**Architecture:** Python MCP server wrapping Obsidian Local REST API HTTP client

**Python + MCP focus:**
- Distributed via PyPI as `mcp-obsidian` package
- Installed with `uv` or `pip`
- Pure HTTP client — no direct plugin dependency (uses REST API over HTTPS)
- Implements MCP server protocol (SSI — Standard Server Interface)

**Key philosophy:**
- Simple, thin wrapper with no reimplementation of vault logic
- All vault operations delegated to obsidian-local-rest-api REST API
- Configuration via environment variables or `.env` file
- Works anywhere Python is available

## Architecture

**Core structure:**
- `mcp-obsidian` — CLI entry point (distributed to PyPI)
- HTTP client → obsidian-local-rest-api API (running in Obsidian plugin)
- MCP server protocol → client (Claude, other agents)

**Tool mapping:**
- MCP tools map 1:1 to obsidian-local-rest-api REST endpoints
- Example: `vault_read` tool → `GET /vault/{path}`
- Search tools → `POST /search/simple/` or `POST /search/`

**Configuration:**
- Obsidian host/port (default 127.0.0.1:27124)
- API key (from Obsidian Local REST API settings)
- Certificate validation (can skip for dev self-signed cert)

## Style

**Code patterns:**
- Async/await for HTTP operations
- Type hints throughout
- Minimal error handling (delegate to REST API)
- Configuration via environment variables

**MCP conventions:**
- Standard SSI (Server-Sent Instructions) format
- Tool definitions that mirror REST API capabilities
- Streaming responses for search results

**Documentation:**
- README with installation and configuration instructions
- Example configs for Claude Desktop and Claude Code
- MCP Inspector debugging instructions

## Testing

### Unit Tests

No dedicated unit tests. Testing relies on integration tests (see below).

### Integration Tests

Manually test the server:

1. **Start obsidian-local-rest-api:**
   - Ensure Obsidian is running with Local REST API plugin enabled
   - Get API key from Settings → Local REST API

2. **Start mcp-obsidian server:**
   ```bash
   OBSIDIAN_API_KEY=<your-key> mcp-obsidian
   ```

3. **Test with MCP Inspector:**
   ```bash
   npx @modelcontextprotocol/inspector uv --directory /path/to/mcp-obsidian run mcp-obsidian
   ```

4. **Verify tools in inspector:**
   - `vault_list` — list vault root
   - `vault_read` — read a note
   - `vault_write` — create/update a note
   - `vault_patch` — patch a section
   - `search_simple` — search notes
   - Other tools matching REST API endpoints

### Build

```bash
uv sync
```

Synchronizes dependencies and updates lockfile for publication.

### Smoke Tests

Manual verification checklist:
1. Ensure Obsidian Local REST API plugin is running and API key is available
2. Start mcp-obsidian with correct OBSIDIAN_API_KEY and OBSIDIAN_HOST
3. Connect Claude Code: `claude mcp add uv --directory /path/to/mcp-obsidian run mcp-obsidian`
4. In Claude Code, test vault operations:
   - Read a note
   - Create a new note
   - Search the vault
   - Patch a heading or section
5. Verify server logs show successful REST API calls

### Subagent Testing

When making changes to tool definitions or MCP protocol compliance:
1. Use MCP Inspector to verify tool signatures
2. Test with Claude agents to ensure tool use works end-to-end
3. Verify response formats match MCP expectations
4. Document any tool changes in CHANGELOG
