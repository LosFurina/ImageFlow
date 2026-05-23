## Why

ImageFlow 当前只有一套 Internal API（`/api/*`），使用单一 Bearer Token 认证，供 WebUI 专用。这带来三个问题：

1. **无法安全开放给外部用户**：单一 API_KEY 全局共享，无法区分用户，无法追踪调用来源，泄露后影响全局
2. **无法细粒度控制权限**：所有持有 API Key 的用户拥有完全相同的权限，无法限制某些用户只能读取、某些用户只能上传
3. **缺少标准化的 API 文档**：外部开发者需要阅读源码才能集成，没有 OpenAPI 规范文档，无法在线调试

## What Changes

- **新增 OpenAPI 端点层**（`/openapi/*`）：与 `/api/*` 完全隔离，复用核心 handler 逻辑但使用 AK/SK 认证
- **AK/SK 认证体系**：简化版 HMAC 签名（AK Header + SK 签名 + 时间戳防重放），存储在 Redis
- **RBAC + 自定义权限**：预设 reader/writer/admin 角色 + 端点粒度自定义权限
- **AK/SK 管理 API**（`/api/admin/aksk/*`）：CRUD + SK 轮换，Bearer Token 认证
- **Swagger 自动生成**：`swaggo/swag` 注解驱动，`/openapi/docs` 嵌入 Swagger UI
- **前端 AK/SK 管理界面**：`/manage` 页面新增管理 Tab

## Capabilities

### New Capabilities
- `openapi-aksk-auth`: AK/SK 认证体系，包含 HMAC 签名验证、Redis 存储、权限校验
- `openapi-endpoints`: `/openapi/*` 端点层，映射所有 Internal API，使用 AK/SK 认证
- `aksk-management`: AK/SK 生命周期管理 API（创建/更新/删除/轮换）
- `swagger-docs`: 自动生成的 OpenAPI 3.0 规范文档和 Swagger UI
- `aksk-frontend`: 前端 AK/SK 管理界面

### Modified Capabilities
- `architecture`: 新增 auth/ 包，路由注册扩展，前端 /manage 页面扩展

## Impact

- **后端新增**: `auth/` 包（4 文件）、`handlers/openapi.go`、`handlers/aksk_admin.go`、`docs/`（自动生成）
- **配置扩展**: `AKSK_ENABLED` 环境变量
- **依赖新增**: `swaggo/swag/v2`、`swaggo/http-swagger/v2`
- **前端修改**: `/manage` 页面新增 Tab、`AKSKManager.tsx` 组件
- **现有 Internal API 不变**: `/api/*` 路由、Bearer Token 认证、handler 逻辑完全不动
- **Redis 新增 key 模式**: `imageflow:aksk:{ak}`
