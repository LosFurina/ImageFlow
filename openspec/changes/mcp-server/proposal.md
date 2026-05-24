# Proposal: MCP Server for ImageFlow

## Why

Model Context Protocol (MCP) is becoming the standard way for AI assistants to connect to external tools and data. By exposing ImageFlow's OpenAPI capabilities as an MCP Server, users can connect Claude Desktop, Cursor, and other MCP-compatible clients directly to their ImageFlow instance — no manual API calls, no HMAC signing, just natural language interaction.

## What Changes

- Add a new `mcp/` package implementing an MCP Server over HTTP SSE
- Map all 8 existing `/openapi/*` endpoints to MCP Tools with descriptive schemas
- Reuse existing AK/SK authentication (validated at SSE connection time)
- Add `MCP_ENABLED` env var and integrate into main.go route registration
- Update README with MCP connection instructions

## Capabilities

### New
- `mcp-server`: MCP Server over HTTP SSE at `/mcp/sse` and `/mcp/messages`
- `mcp-tools`: 8 MCP Tools mapped from OpenAPI endpoints:
  - `imageflow_upload` — Upload images with tags and expiry
  - `imageflow_list_images` — List uploaded images with pagination and tag filters
  - `imageflow_delete_image` — Delete an image by ID
  - `imageflow_list_tags` — List all available tags
  - `imageflow_get_config` — Get server configuration
  - `imageflow_get_random_image` — Get a random image with filters
  - `imageflow_debug_tags` — Debug tag index data
  - `imageflow_trigger_cleanup` — Trigger expired image cleanup

### Modified
- `main.go`: Register MCP routes when `MCP_ENABLED=true` and `AKSK_ENABLED=true`
- `config.go`: Add `MCP_ENABLED` configuration field
- `README.md` / `README_CN.md`: Add MCP setup section

## Impact

- No breaking changes to existing `/api/*` or `/openapi/*` endpoints
- MCP is opt-in via `MCP_ENABLED=true`
- Requires `AKSK_ENABLED=true` and Redis (same as OpenAPI)
- Adds one new dependency: `github.com/mark3labs/mcp-go`
