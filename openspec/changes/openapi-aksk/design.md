## Context

ImageFlow currently has a single API layer (`/api/*`) protected by one global Bearer Token (`API_KEY` env var). This is sufficient for the WebUI but cannot serve external users or scripts safely. There's no user identity, no permission granularity, and no standardized API documentation.

The project uses Go 1.23+ with `net/http` standard library (no framework), Redis for metadata, and Next.js 15 for the frontend.

## Goals / Non-Goals

**Goals:**
- Add a fully isolated `/openapi/*` endpoint layer with AK/SK HMAC authentication
- Implement RBAC with 3 preset roles + custom per-user permissions
- Store AK/SK data in Redis (consistent with existing metadata storage)
- Provide admin CRUD APIs for AK/SK management (Bearer Token protected)
- Auto-generate OpenAPI 3.0 spec via `swaggo/swag` annotations
- Embed Swagger UI at `/openapi/docs`
- Add basic AK/SK management tab in the frontend `/manage` page
- Keep all existing `/api/*` routes and Bearer Token auth completely untouched

**Non-Goals:**
- No changes to existing Internal API behavior or authentication
- No rate limiting (can be added later)
- No audit logging beyond standard zap logs (can be added later)
- No fancy UI — basic functional interface only
- No webhook/callback support
- No multi-tenant or organization-level isolation
- No OAuth/JWT integration

## Decisions

1. **Swaggo/swag for spec generation**: Industry-standard Go library. Annotations live with code, spec stays in sync. `swag init` generates `docs/`. `swaggo/http-swagger` serves Swagger UI.

2. **Separate `auth/` package**: New top-level package for AK/SK logic, cleanly separated from existing `handlers/auth.go`. This avoids tangling new auth code with the existing Bearer Token middleware.

3. **Handler wrapping pattern**: OpenAPI handlers wrap the same core logic as Internal API handlers but go through `AKSKAuthMiddleware` instead of `RequireAPIKey`. This avoids code duplication while keeping auth layers isolated.

4. **SK stored encrypted with AES-GCM**: Redis never stores plaintext SK. The SK is encrypted with AES-GCM using a key derived from the existing API_KEY, allowing the server to decrypt it for HMAC verification while protecting Redis at rest. Lost SKs still require rotation because the UI never exposes existing SKs.

5. **Redis for AK/SK storage**: Consistent with existing metadata approach. AK/SK data is simple key-value, Redis is already required and initialized. No new infrastructure.

6. **AK/SK format**: AK = 20-char base62, SK = 40-char base62. Long enough to resist brute force, short enough to be practical.

7. **5-minute timestamp tolerance**: Standard anti-replay window. Balances security against clock skew between clients and server.

8. **Permission resolution**: `effective_permissions = role_permissions ∪ custom_permissions`. Simple union — custom permissions only add, never subtract. This keeps the mental model simple.

9. **`AKSK_ENABLED` config flag**: OpenAPI is opt-in. When disabled, no OpenAPI routes are registered and no Redis AK/SK queries happen. Existing deployments aren't affected until they enable it.

10. **Front-end placement**: AK/SK tab in existing `/manage` page. Uses existing layout, auth flow, and API client. No new pages or routes needed.

## Risks / Trade-offs

- **SK recovery impossible**: Since existing SKs are never exposed through the UI, lost SKs require rotation. Redis stores encrypted SK values rather than plaintext. Mitigation: clear UI warning when SK is shown.

- **Redis single point of failure**: If Redis goes down, AK/SK auth fails → all OpenAPI requests return 503. This is the same risk as existing metadata storage. Mitigation: Redis persistence (RDB/AOF), same as current setup.

- **Handler wrapping complexity**: OpenAPI handlers wrapping Internal handlers adds an indirection layer. If Internal handler signatures change, OpenAPI wrappers must be updated too. Mitigation: keep wrapper thin, just auth + delegate.

- **swaggo annotation maintenance**: Annotations must be kept in sync with code. If someone changes a handler without updating annotations, the spec becomes stale. Mitigation: CI validation (future), developer discipline.

- **No backward compatibility for SK format**: If we later want to change the signature algorithm or SK format, existing AK/SK pairs need migration. Mitigation: version field in AK/SK data for future algorithm upgrades.

- **Frontend minimal scope**: The management UI will be functional but not polished. May need iteration based on user feedback.
