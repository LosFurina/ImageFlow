## Purpose
Expose ImageFlow OpenAPI capabilities as an MCP Server over HTTP SSE, allowing MCP-compatible clients to discover and invoke image management tools via the Model Context Protocol.

## Requirements

### Requirement: mcp-server-enable
The system SHALL expose an MCP Server endpoint when `MCP_ENABLED=true` and `AKSK_ENABLED=true`.

#### Scenario: MCP enabled with prerequisites
- **WHEN** `MCP_ENABLED=true`, `AKSK_ENABLED=true`, and Redis is available
- **THEN** the system registers MCP routes and starts accepting MCP connections

#### Scenario: MCP disabled
- **WHEN** `MCP_ENABLED=false`
- **THEN** no MCP routes are registered and MCP connections are rejected

#### Scenario: MCP enabled but AK/SK disabled
- **WHEN** `MCP_ENABLED=true` but `AKSK_ENABLED=false`
- **THEN** the system logs a warning and does not register MCP routes

### Requirement: mcp-transport-sse
The system SHALL support MCP over HTTP SSE transport at `/mcp/sse` and `/mcp/messages`.

#### Scenario: Client establishes SSE connection
- **WHEN** an MCP client sends `GET /mcp/sse` with valid AK/SK headers
- **THEN** the server establishes an SSE connection and returns an `endpoint` event containing the message URL

#### Scenario: Client sends MCP message
- **WHEN** an MCP client sends `POST /mcp/messages` with a valid JSON-RPC message
- **THEN** the server processes the message and sends the response via the SSE stream

#### Scenario: Invalid AK/SK on SSE connection
- **WHEN** a client sends `GET /mcp/sse` without valid AK/SK headers
- **THEN** the connection is rejected with HTTP 401

### Requirement: mcp-tool-discovery
The system SHALL expose all OpenAPI capabilities as discoverable MCP Tools via `tools/list`.

#### Scenario: List available tools
- **WHEN** an MCP client sends `tools/list`
- **THEN** the server returns all 8 ImageFlow tools with name, description, and JSON Schema input parameters

#### Scenario: Tool name format
- **WHEN** inspecting tool names returned by `tools/list`
- **THEN** each tool name uses snake_case prefixed with `imageflow_`

### Requirement: mcp-tool-invoke
The system SHALL allow MCP clients to invoke OpenAPI operations via `tools/call`.

#### Scenario: Invoke upload tool
- **WHEN** an MCP client calls `imageflow_upload` with valid parameters
- **THEN** the server executes the upload operation and returns the result as MCP content

#### Scenario: Invoke list images tool
- **WHEN** an MCP client calls `imageflow_list_images` with page and tag filters
- **THEN** the server returns the paginated image list as MCP text content

#### Scenario: Tool returns error
- **WHEN** a tool invocation fails (e.g., permission denied, invalid parameters)
- **THEN** the server returns an MCP `TextContent` with the error message and sets `isError=true`

### Requirement: mcp-auth-reuse
The system SHALL reuse existing AK/SK authentication for MCP connections.

#### Scenario: Valid AK/SK credentials
- **WHEN** a client connects with valid `X-Access-Key`, `X-Signature`, and `X-Timestamp` headers
- **THEN** the MCP session is established and all tool invocations use those credentials

#### Scenario: Permission check on tool invoke
- **WHEN** an AK/SK user with `reader` role invokes `imageflow_upload` (requires `api:upload`)
- **THEN** the invocation is rejected with a permission error

### Requirement: mcp-isolation
The system SHALL keep MCP endpoints isolated from existing `/api/*` and `/openapi/*` endpoints.

#### Scenario: No interference with existing APIs
- **WHEN** MCP routes are registered
- **THEN** `/api/*` and `/openapi/*` endpoints continue to work exactly as before
