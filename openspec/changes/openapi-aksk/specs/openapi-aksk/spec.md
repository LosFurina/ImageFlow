## Purpose

Define the requirements for ImageFlow's OpenAPI layer with AK/SK authentication, RBAC permissions, management APIs, Swagger documentation, and frontend management interface.

## Requirements

### Requirement: aksk-authentication
The system SHALL authenticate OpenAPI requests using Access Key / Secret Key with HMAC-SHA256 signature verification.

#### Scenario: Valid AK/SK request
- **WHEN** a request includes headers `X-Access-Key`, `X-Signature`, `X-Timestamp` and the signature matches `HMAC-SHA256(SK, Method + Path + Timestamp + SHA256(Body))`
- **THEN** the request is authenticated and proceeds to permission checking

#### Scenario: Invalid signature
- **WHEN** a request's computed signature does not match the `X-Signature` header
- **THEN** the system returns HTTP 401 with error code 1002

#### Scenario: Expired timestamp
- **WHEN** a request's `X-Timestamp` differs from server time by more than 5 minutes
- **THEN** the system returns HTTP 401 with error code 1002 and message "Request timestamp expired"

#### Scenario: Disabled AK
- **WHEN** a request uses an AK that has `enabled: false`
- **THEN** the system returns HTTP 403 with error code 1003

### Requirement: aksk-storage
The system SHALL store AK/SK credentials and metadata in Redis with SK stored AES-GCM encrypted, never in plaintext.

#### Scenario: Create AK/SK pair
- **WHEN** an admin creates a new AK/SK pair
- **THEN** the system generates a random AK (20 chars) and SK (40 chars), stores AES-GCM encrypted SK in Redis under key `imageflow:aksk:{ak}`, and returns the plaintext AK/SK to the admin once

#### Scenario: SK never stored plaintext
- **WHEN** examining Redis data
- **THEN** no SK plaintext exists; only AES-GCM encrypted SK is stored

#### Scenario: SK rotation
- **WHEN** an admin rotates SK for an AK
- **THEN** a new SK is generated, the old encrypted SK is replaced, and the new plaintext SK is returned once

### Requirement: rbac-permissions
The system SHALL enforce role-based access control with custom permission override on every OpenAPI request.

#### Scenario: Role-based access
- **WHEN** an AK/SK user with role `reader` requests `POST /openapi/upload` (requires `api:upload`)
- **THEN** the system returns HTTP 403 with error code 1003

#### Scenario: Custom permission grants access
- **WHEN** an AK/SK user with role `reader` has custom permission `api:upload` added
- **THEN** the user can access `POST /openapi/upload`

#### Scenario: Effective permissions
- **WHEN** computing a user's effective permissions
- **THEN** the result is the union of role permissions and custom permissions

#### Scenario: Admin role has all permissions
- **WHEN** an AK/SK user with role `admin` requests any OpenAPI endpoint
- **THEN** access is always granted

### Requirement: openapi-endpoints
The system SHALL expose all Internal API capabilities under `/openapi/*` prefix with AK/SK authentication, reusing the same handler logic.

#### Scenario: OpenAPI mirrors Internal API
- **WHEN** comparing `/openapi/*` endpoints with `/api/*` endpoints
- **THEN** every authenticated Internal API has a corresponding OpenAPI endpoint with the same request/response format

#### Scenario: OpenAPI uses AK/SK only
- **WHEN** a request to `/openapi/*` includes Bearer Token instead of AK/SK headers
- **THEN** the system returns HTTP 401

#### Scenario: Internal API unaffected
- **WHEN** a request to `/api/*` includes AK/SK headers instead of Bearer Token
- **THEN** the system returns HTTP 401 (existing behavior unchanged)

### Requirement: aksk-management-api
The system SHALL provide admin APIs for AK/SK lifecycle management, authenticated via existing Bearer Token.

#### Scenario: Create AK/SK
- **WHEN** an admin sends `POST /api/admin/aksk/create` with `name`, `role`, and optional `custom_permissions`
- **THEN** the system creates a new AK/SK pair and returns the AK and plaintext SK

#### Scenario: List AK/SK
- **WHEN** an admin sends `GET /api/admin/aksk/list`
- **THEN** the system returns all AK/SK entries with AK, name, role, permissions, enabled status, created_at (SK excluded)

#### Scenario: Update AK/SK
- **WHEN** an admin sends `PUT /api/admin/aksk/update` with AK and fields to change
- **THEN** the system updates the specified fields (role, custom_permissions, enabled, name, description)

#### Scenario: Delete AK/SK
- **WHEN** an admin sends `DELETE /api/admin/aksk/delete` with AK
- **THEN** the system removes the AK/SK entry from Redis

#### Scenario: Rotate SK
- **WHEN** an admin sends `POST /api/admin/aksk/rotate` with AK
- **THEN** the system generates a new SK, replaces the hash, and returns the new plaintext SK

### Requirement: swagger-docs
The system SHALL auto-generate OpenAPI 3.0 specification from Go handler annotations and serve an interactive Swagger UI.

#### Scenario: Access Swagger UI
- **WHEN** a user navigates to `/openapi/docs`
- **THEN** the system serves an interactive Swagger UI with all OpenAPI endpoints documented

#### Scenario: Auto-generated spec
- **WHEN** `swag init` is run after updating handler annotations
- **THEN** the system generates `docs/swagger.json` and `docs/swagger.yaml` reflecting all documented endpoints

#### Scenario: Spec matches implementation
- **WHEN** comparing the generated spec with actual OpenAPI endpoints
- **THEN** every endpoint, parameter, and response model is accurately documented

### Requirement: aksk-frontend
The system SHALL provide a frontend management interface for AK/SK in the `/manage` page.

#### Scenario: View AK/SK list
- **WHEN** a user navigates to the AK/SK tab in `/manage`
- **THEN** the system displays a table with AK, name, role, status, and created time for all entries

#### Scenario: Create AK/SK via UI
- **WHEN** a user clicks "Create" and fills in name and role
- **THEN** the system creates the AK/SK and displays a one-time dialog showing the SK with a warning to save it

#### Scenario: SK shown only once
- **WHEN** a new AK/SK is created
- **THEN** the SK is displayed in a modal exactly once and cannot be retrieved again after closing

#### Scenario: Manage AK/SK
- **WHEN** a user toggles enabled/disabled or clicks delete on an AK/SK entry
- **THEN** the system updates or removes the entry accordingly with confirmation
