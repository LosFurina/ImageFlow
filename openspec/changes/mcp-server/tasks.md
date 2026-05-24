# Tasks: MCP Server Implementation

## Phase 1: Foundation

- [ ] 1.1 Add `github.com/mark3labs/mcp-go` to go.mod
  - `go get github.com/mark3labs/mcp-go`
  - `go mod tidy`

- [ ] 1.2 Add `MCP_ENABLED` to config
  - Add `MCPEnabled bool` to `config.Config`
  - Load from env var `MCP_ENABLED` (default: false)
  - Add to `.env.example`

- [ ] 1.3 Create `mcp/auth.go`
  - `ValidateAKSKForMCP(r *http.Request, rdb *redis.Client) (*auth.AKSKCredentials, error)`
  - Reuses existing `auth.VerifySignature` and `auth.ValidateAKSKRequest`
  - Returns credentials + permissions for session storage

## Phase 2: MCP Server Core

- [ ] 2.1 Create `mcp/tools.go`
  - Define 8 `mcp.Tool` structs with name, description, JSON Schema input
  - `imageflow_upload`: params `images` (array of base64 strings), `tags` (string), `expiryMinutes` (int)
  - `imageflow_list_images`: params `page` (int), `pageSize` (int), `tag` (string)
  - `imageflow_delete_image`: params `id` (string)
  - `imageflow_list_tags`: no params
  - `imageflow_get_config`: no params
  - `imageflow_get_random_image`: params `tag` (string), `tags` (string), `exclude` (string), `orientation` (string), `format` (string)
  - `imageflow_debug_tags`: no params
  - `imageflow_trigger_cleanup`: no params

- [ ] 2.2 Create `mcp/handlers.go`
  - `handleUpload(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleListImages(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleDeleteImage(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleListTags(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleGetConfig(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleGetRandomImage(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleDebugTags(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - `handleTriggerCleanup(ctx, request, rdb, cfg) → *mcp.CallToolResult`
  - Each handler reuses existing handler logic, returns JSON-serialized result as `mcp.TextContent`

- [ ] 2.3 Create `mcp/server.go`
  - `NewServer(rdb *redis.Client, cfg *config.Config) *server.MCPServer`
  - Create `MCPServer` with name "ImageFlow" and version from build
  - Register all 8 tools with handlers
  - Return configured server

## Phase 3: HTTP Integration

- [ ] 3.1 Create `mcp/transport.go`
  - `RegisterMCPRoutes(rdb *redis.Client, cfg *config.Config)`
  - Wraps SSE transport with AK/SK auth middleware
  - Registers `GET /mcp/sse` (with AK/SK validation)
  - Registers `POST /mcp/messages` (MCP message endpoint)

- [ ] 3.2 Integrate into `main.go`
  - After OpenAPI route registration, add:
    ```go
    if cfg.MCPEnabled && cfg.AKSKEnabled && rdb != nil {
        mcp.RegisterMCPRoutes(rdb, cfg)
        logger.Info("MCP Server enabled at /mcp/sse")
    }
    ```
  - Log warning if `MCP_ENABLED=true` but prerequisites not met

## Phase 4: Testing & Validation

- [ ] 4.1 Build check
  - `go build ./...` passes
  - No compilation errors

- [ ] 4.2 Manual test with MCP inspector or curl
  - `curl` to `/mcp/sse` with valid AK/SK headers → 200 + SSE stream
  - `curl` to `/mcp/sse` without headers → 401
  - Send `tools/list` message → returns 8 tools
  - Send `tools/call` for `imageflow_list_tags` → returns tags

- [ ] 4.3 Verify existing endpoints untouched
  - `/api/*` still works with Bearer Token
  - `/openapi/*` still works with AK/SK
  - `/openapi/docs` still serves Swagger UI

## Phase 5: Documentation

- [ ] 5.1 Update `README.md` and `README_CN.md`
  - Add MCP section: what it is, how to enable, how to connect
  - Include `claude_desktop_config.json` example:
    ```json
    {
      "mcpServers": {
        "imageflow": {
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-sse"],
          "env": {
            "SSE_URL": "http://localhost:8686/mcp/sse"
          }
        }
      }
    }
    ```
  - Note: MCP requires `MCP_ENABLED=true` and `AKSK_ENABLED=true`

- [ ] 5.2 Update `.env.example`
  - Add `MCP_ENABLED=false`

- [ ] 5.3 Update `CLAUDE.md`
  - Add `mcp/` package to project structure
  - Add MCP dev commands
