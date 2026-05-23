## Phase 1: Infrastructure (auth/ package)

- [x] 1.1 Create `auth/permissions.go`: define permission constants (`api:random`, `api:upload`, `api:images`, `api:delete`, `api:tags`, `api:config`, `api:debug`, `api:cleanup`), role presets (`reader`, `writer`, `admin`), `EffectivePermissions(role, custom) []string` function
- [x] 1.2 Create `auth/aksk.go`: `AKSKEntry` struct, `GenerateAK()` / `GenerateSK()` functions, Redis CRUD (Save, Load, Delete, List), SK encryption (`AES-GCM`, no plaintext Redis storage)
- [x] 1.3 Create `auth/signature.go`: `BuildStringToSign(method, path, timestamp, bodyHash)`, `ComputeSignature(sk, stringToSign)`, `VerifySignature(sk, method, path, timestamp, body, providedSig)`, `SHA256Body(body)`, `ValidateTimestamp(ts) error`
- [x] 1.4 Create `auth/middleware.go`: `AKSKAuthMiddleware(requiredPermission string, next http.HandlerFunc) http.HandlerFunc` — extracts AK/SK headers, looks up Redis, verifies signature, validates timestamp, checks permission, calls next
- [x] 1.5 Add `AKSK_ENABLED` field to `config.Config`, load from env var, skip AK/SK routes when false
- [x] 1.6 Write unit tests for `auth/` package: signature verification, permission resolution, AK/SK generation

## Phase 2: OpenAPI Routes

- [x] 2.1 Create `handlers/openapi.go`: register `/openapi/*` routes using `AKSKAuthMiddleware` wrapping existing handler logic
  - `POST /openapi/upload` → `api:upload` → wraps upload logic
  - `GET /openapi/images` → `api:images` → wraps list logic
  - `POST /openapi/delete` → `api:delete` → wraps delete logic
  - `GET /openapi/tags` → `api:tags` → wraps tags logic
  - `GET /openapi/config` → `api:config` → wraps config logic
  - `GET /openapi/random` → `api:random` → wraps random logic
  - `GET /openapi/debug/tags` → `api:debug` → wraps debug tags logic
  - `POST /openapi/cleanup` → `api:cleanup` → wraps cleanup logic
- [x] 2.2 Add swag annotations to each OpenAPI handler
- [x] 2.3 Register OpenAPI routes in `main.go` (conditionally when `AKSK_ENABLED=true`)
- [x] 2.4 Install `swaggo/swag` and `swaggo/http-swagger`, run `swag init`, verify spec generation

## Phase 3: Management API

- [x] 3.1 Create `handlers/aksk_admin.go`: admin CRUD handlers with swag annotations
  - `GET /api/admin/aksk/list` → list all AK/SK entries (no SK returned)
  - `POST /api/admin/aksk/create` → create AK/SK with name, role, optional custom_permissions
  - `PUT /api/admin/aksk/update` → update role, custom_permissions, enabled, name, description
  - `DELETE /api/admin/aksk/delete` → delete AK/SK entry
  - `POST /api/admin/aksk/rotate` → generate new SK, return plaintext once
- [x] 3.2 Wrap admin handlers with `RequireAPIKey` (Bearer Token auth)
- [x] 3.3 Register admin routes in `main.go`
- [x] 3.4 Test admin API: create, list, update, rotate, delete cycle

## Phase 4: Swagger UI

- [x] 4.1 Add `swaggo/http-swagger` handler for `/openapi/docs` and `/openapi/docs/*`
- [x] 4.2 Configure Swagger UI with correct API base path and security scheme display
- [x] 4.3 Add AK/SK security scheme definition in swag annotations (`@SecurityDefinitions.apikey AkSkAuth`)
- [x] 4.4 Verify Swagger UI loads and all endpoints are documented and testable
- [x] 4.5 Add `swag init` instruction to README

## Phase 5: Frontend

- [x] 5.1 Create `frontend/app/components/AKSKManager.tsx`: main AK/SK management component
  - Table: AK (masked), name, role badge, enabled toggle, created time, actions
  - Create dialog: name input, role select, optional custom permissions checkboxes
  - SK display modal: one-time show with copy button and warning
  - Delete confirmation dialog
  - Rotate SK dialog (shows new SK once)
- [x] 5.2 Add AK/SK tab to `/manage` page (`frontend/app/manage/page.tsx`)
- [x] 5.3 Add frontend API calls: `listAKSK`, `createAKSK`, `updateAKSK`, `deleteAKSK`, `rotateSK`
- [x] 5.4 Basic styling with Tailwind (consistent with existing UI, no fancy design)

## Phase 6: Integration & Validation

- [x] 6.1 End-to-end smoke test: create AK/SK → sign request → call OpenAPI config → verify permission rejection
- [x] 6.2 Permission test: verify each role can only access permitted endpoints
- [x] 6.3 Timestamp expiry test: verify expired requests are rejected
- [x] 6.4 Isolation test: verify `/api/*` still works with Bearer Token, verify `/openapi/*` rejects Bearer Token
- [x] 6.5 `go build ./...` and `go vet ./...` pass
- [x] 6.6 Update README.md and README_CN.md with OpenAPI documentation section
- [x] 6.7 Update CLAUDE.md with new architecture information
- [x] 6.8 Update openspec/specs/architecture/ with new capabilities
- [x] 6.9 Commit and push
