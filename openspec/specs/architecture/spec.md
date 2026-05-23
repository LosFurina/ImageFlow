## Purpose

ImageFlow is a modern image service system that provides efficient image management and distribution. It automatically optimizes images based on device type and browser compatibility, supporting WebP/AVIF conversion, tag-based categorization, and flexible storage backends.

## Requirements

### Requirement: image-upload-processing
The system SHALL accept image uploads via HTTP POST, detect image format and orientation, store originals, and asynchronously convert to WebP/AVIF using a worker pool with configurable concurrency.

#### Scenario: Upload a JPEG image
- **WHEN** a user submits a JPEG file to `POST /api/upload` with a valid API key
- **THEN** the system stores the original image, determines orientation (landscape if width > height, portrait otherwise), enqueues WebP and AVIF conversions to the worker pool, and returns upload results including URLs for all formats

#### Scenario: Upload a GIF image
- **WHEN** a user submits a GIF file
- **THEN** the system stores it without conversion to preserve animation, and returns URLs pointing to the gif storage path

#### Scenario: Upload with tags and expiry
- **WHEN** a user uploads with optional tags and expiry minutes
- **THEN** the system stores metadata with tags and expiry timestamp for later cleanup

### Requirement: image-serving-random
The system SHALL serve random images based on device type, tags, orientation, and format preferences, with browser-compatible format negotiation.

#### Scenario: Mobile device requests random image
- **WHEN** a request comes from a mobile User-Agent to `GET /api/random`
- **THEN** the system defaults to portrait orientation images and serves AVIF > WebP > Original based on Accept headers

#### Scenario: Filter by tags with exclusion
- **WHEN** a request includes `tags=nature,sunset&exclude=nsfw`
- **THEN** the system returns a random image matching ALL specified tags, excluding any with the listed exclusion tags

### Requirement: storage-abstraction
The system SHALL abstract storage operations behind a provider interface supporting local filesystem and S3-compatible backends.

#### Scenario: Local storage
- **WHEN** `STORAGE_TYPE=local`
- **THEN** images are stored in a directory structure: `{base}/{orientation}/{format}/` with originals in `{base}/original/{orientation}/`

#### Scenario: S3 storage
- **WHEN** `STORAGE_TYPE=s3` with valid S3 credentials
- **THEN** images are stored in the configured S3 bucket with the same key structure, and public URLs use custom domain or endpoint-based paths

### Requirement: metadata-management
The system SHALL manage image metadata with Redis as primary store and file-based fallback.

#### Scenario: Redis available
- **WHEN** Redis is configured and reachable
- **THEN** metadata is stored in Redis with key prefix `imageflow:`, supporting tag-indexed queries, page caching (5 min TTL), and expiry-based cleanup

#### Scenario: Redis unavailable
- **WHEN** Redis connection fails
- **THEN** the system falls back to JSON file-based metadata storage in `{base}/metadata/`

### Requirement: api-key-auth
The system SHALL authenticate upload, delete, and management endpoints using API key via Bearer token with constant-time comparison.

#### Scenario: Valid API key
- **WHEN** a request includes `Authorization: Bearer ***
- **THEN** the request is processed normally

#### Scenario: Invalid API key
- **WHEN** a request includes an invalid key
- **THEN** the system returns error code 1002 (Unauthorized) with no timing leakage

### Requirement: image-cleanup
The system SHALL periodically clean up expired images based on their expiry metadata.

#### Scenario: Image with expiry time passed
- **WHEN** the cleanup interval triggers and an image's expiry time has passed
- **THEN** the system deletes the image from storage (local or S3), removes metadata, and removes tag associations

### Requirement: frontend-upload-ui
The frontend SHALL provide a Next.js SPA with drag-and-drop upload, tag selection, expiry settings, image preview, and dark/light theme support.

#### Scenario: Upload flow
- **WHEN** a user drags images into the dropzone
- **THEN** a preview sidebar shows selected files, the user can set tags and expiry, and clicking upload sends files to the backend with the API key from localStorage

#### Scenario: Image management page
- **WHEN** a user navigates to `/manage`
- **THEN** the system displays a paginated masonry grid of all images with filtering by format, orientation, and tags, and supports deletion with confirmation


### Requirement: openapi-aksk-access
The system SHALL expose external script-facing OpenAPI endpoints under `/openapi/*` protected by AK/SK HMAC authentication while leaving existing `/api/*` Bearer Token internal APIs unchanged.

#### Scenario: External script uses AK/SK
- **WHEN** a script calls `/openapi/upload` with valid `X-Access-Key`, `X-Signature`, and `X-Timestamp` headers
- **THEN** the system validates the HMAC signature, checks the required permission, and processes the request

#### Scenario: Internal API remains isolated
- **WHEN** WebUI calls `/api/*` endpoints
- **THEN** the system continues to use the existing Bearer Token API key flow without requiring AK/SK headers

### Requirement: aksk-management-ui
The system SHALL provide basic AK/SK lifecycle management in the frontend `/manage` page for creating, disabling, rotating, and deleting external API credentials.

#### Scenario: Create credential from UI
- **WHEN** an authenticated admin creates an AK/SK credential in `/manage`
- **THEN** the system returns the Secret Key exactly once and lists the Access Key with role and enabled status

### Requirement: swagger-documentation
The system SHALL generate Swagger documentation from Go annotations and serve it at `/openapi/docs` when AK/SK OpenAPI is enabled.

#### Scenario: Swagger UI access
- **WHEN** a user opens `/openapi/docs`
- **THEN** the system serves an interactive Swagger UI containing all `/openapi/*` endpoints
