## Context

ImageFlow already has:
- `/api/*` endpoints with Bearer Token auth (internal WebUI)
- `/openapi/*` endpoints with AK/SK HMAC auth (external scripts)
- 8 OpenAPI operations: upload, list images, delete, tags, config, random, debug tags, cleanup
- Redis-based AK/SK storage with role/permission system
- Go 1.23+ with standard HTTP mux

## Goals

- Expose all OpenAPI capabilities as MCP Tools over HTTP SSE
- Reuse existing AK/SK authentication and permission system
- Zero breaking changes to existing endpoints
- Opt-in via `MCP_ENABLED` env var

## Non-Goals

- Support stdio transport (out of scope for a backend HTTP service)
- Expose `/api/*` internal endpoints as MCP Tools (only OpenAPI)
- Implement MCP Resources or Prompts (Tools only, for now)
- Replace OpenAPI with MCP — they coexist

## Design Decisions

### 1. Transport: HTTP SSE

**Decision**: Use HTTP SSE transport at `/mcp/sse` and `/mcp/messages`.

**Rationale**: ImageFlow is already an HTTP server. SSE is the standard MCP remote transport. stdio would require a separate binary/process model.

### 2. SDK: `mcp-go`

**Decision**: Use `github.com/mark3labs/mcp-go` (official Go SDK for MCP).

**Rationale**: Mature, well-documented, supports SSE transport out of the box. Saves implementing JSON-RPC 2.0 + MCP protocol manually.

### 3. Package Structure

```
mcp/
├── server.go       # MCP server initialization, SSE transport setup
├── tools.go        # Tool definitions (name, description, schema)
├── handlers.go     # Tool invocation handlers (call OpenAPI logic)
└── auth.go         # AK/SK validation wrapper for SSE connections
```

### 4. Authentication Flow

```
Client -> GET /mcp/sse
         Headers: X-Access-Key, X-Signature, X-Timestamp
         -> AK/SKAuthMiddleware validates credentials
         -> If valid: establish SSE session, store AK in session context
         -> If invalid: return 401

Client -> POST /mcp/messages
         Body: JSON-RPC message (tools/list or tools/call)
         -> MCP server processes
         -> For tools/call: use stored AK to re-validate permission for specific tool
         -> Execute tool logic (reuses existing handler functions)
         -> Return result via SSE stream
```

### 5. Tool-to-OpenAPI Mapping

| MCP Tool Name | OpenAPI Endpoint | Permission | Handler Reuse |
|---|---|---|---|
| `imageflow_upload` | POST /openapi/upload | `api:upload` | `UploadHandler` |
| `imageflow_list_images` | GET /openapi/images | `api:images` | `PublicListImagesHandler` |
| `imageflow_delete_image` | POST /openapi/delete | `api:delete` | `DeleteImageHandler` |
| `imageflow_list_tags` | GET /openapi/tags | `api:tags` | `TagsHandler` |
| `imageflow_get_config` | GET /openapi/config | `api:config` | `ConfigHandler` |
| `imageflow_get_random_image` | GET /openapi/random | `api:random` | `LocalRandomImageHandler` / `RandomImageHandler` |
| `imageflow_debug_tags` | GET /openapi/debug/tags | `api:debug` | `DebugTagsHandler` |
| `imageflow_trigger_cleanup` | POST /openapi/cleanup | `api:cleanup` | Inline cleanup trigger |

Each tool defines JSON Schema input parameters matching the OpenAPI endpoint's request parameters.

### 6. Permission Validation

Two-level validation:
1. **Connection level**: AK/SK signature valid (at SSE handshake)
2. **Tool level**: AK has the specific permission required for the tool (at `tools/call`)

This matches the existing OpenAPI middleware behavior.

### 7. Error Handling

- Tool execution errors: return MCP `TextContent` with `isError=true`
- Auth errors: return MCP error response with appropriate code
- Invalid parameters: return MCP error with `InvalidParams` code

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| `mcp-go` dependency adds weight | Single focused dependency (~few hundred KB), widely used |
| SSE connections are long-lived | Go HTTP server handles this fine; no extra goroutine concerns |
| AK/SK headers on every SSE request | Standard HTTP header pattern; same as OpenAPI |
| MCP spec evolves | `mcp-go` tracks the spec; we pin to a stable version |

## Dependencies

- `github.com/mark3labs/mcp-go` — MCP Go SDK (new)
- Existing: `go-redis`, `zap`, etc. (already present)
