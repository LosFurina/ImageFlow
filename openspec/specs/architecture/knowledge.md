# ImageFlow 项目架构知识

## 项目概述
ImageFlow 是一个现代化的图片服务系统，提供高效的图片管理和分发能力。全栈应用：Go 后端 + Next.js 前端。

## 技术栈
- **后端**: Go 1.23+, bimg/libvips (图片处理), AWS SDK v2 (S3), go-redis v9, Zap 日志
- **前端**: Next.js 15.5, React 18, Tailwind CSS 3.4, Framer Motion, Lucide icons, Radix UI
- **部署**: Docker Compose (backend + frontend + redis), 支持 pre-built 镜像和本地构建
- **规范**: OpenSpec v1.3.1 (spec-driven workflow: proposal → specs → design → tasks)

## 后端架构 (Go)
- **入口**: `main.go` — HTTP 服务器, CORS 中间件, 优雅关停
- **路由**: 所有 handler 在 `handlers/` 目录, 返回 `http.HandlerFunc` 闭包
- **存储抽象**: `StorageProvider` 接口 (Store/Get/Delete) — `LocalStorage` 和 `S3Storage` 实现
- **元数据**: `MetadataStore` 接口 — Redis 优先, JSON 文件回退
- **图片处理**: bimg/libvips 转 WebP/AVIF, WorkerPool 并发处理
- **认证**: Bearer Token + constant-time comparison
- **清理**: `ImageCleaner` 定时清理过期图片

## Go 模块路径
已迁移: `github.com/LosFurina/ImageFlow`

## API 端点
- `GET /api/random` — 随机图片 (支持 tag/tags/exclude/orientation/format 参数)
- `POST /api/upload` — 上传图片 (需认证, 支持多文件, tag, expiry)
- `GET /api/images` — 列出图片 (分页, 可按 tag 过滤)
- `POST /api/delete-image` — 删除图片 (需认证)
- `GET /api/tags` — 获取所有标签
- `POST /api/validate-api-key` — 验证 API Key
- `GET /api/config` — 获取客户端配置
- `POST /api/trigger-cleanup` — 手动触发清理

## 前端架构 (Next.js)
- App Router, TypeScript, Tailwind CSS
- 页面: `/` (上传页), `/manage` (管理页)
- 认证: API Key 存 localStorage, 通过 Bearer Token 传递
- API 代理: `/api/config` route 转发后端配置
- 主题: dark/light 切换, localStorage 持久化
- 动态背景动画, 瀑布流布局

## 目录结构
```
├── main.go                    # 后端入口
├── config/config.go           # 配置加载
├── handlers/                  # HTTP 处理器
│   ├── auth.go               # API Key 验证
│   ├── upload.go             # 上传处理
│   ├── random.go             # 随机图片
│   ├── list.go               # 图片列表 (分页+缓存)
│   ├── delete.go             # 删除处理
│   ├── tags.go               # 标签管理
│   ├── config_handler.go     # 配置端点
│   └── debug.go              # 调试工具
├── utils/                     # 工具包
│   ├── storage.go            # 存储抽象 (Local/S3)
│   ├── redis.go              # Redis 元数据操作
│   ├── metadata.go           # 元数据模型和文件回退
│   ├── converter_bimg.go     # libvips 图片转换
│   ├── worker_pool.go        # 并发工作池
│   ├── cleaner.go            # 过期图片清理
│   ├── device.go             # 设备检测 (UA 解析)
│   ├── image.go              # 图片格式检测
│   ├── s3client.go           # S3 客户端初始化
│   ├── helpers.go            # IO 辅助
│   ├── logger/logger.go      # Zap 日志
│   └── errors/errors.go      # 错误码体系
├── frontend/                  # Next.js 前端
│   ├── app/                  # App Router
│   │   ├── page.tsx         # 上传页
│   │   ├── manage/page.tsx  # 管理页
│   │   ├── components/      # React 组件
│   │   ├── hooks/           # useConfig, useTheme
│   │   ├── utils/           # request, auth, baseUrl
│   │   └── types/           # TypeScript 类型
│   └── next.config.mjs      # Next.js 配置
├── scripts/convert.go        # 批量转换工具
├── docker-compose.yaml       # 生产部署
├── docker-compose.build.yaml # 开发构建
├── Dockerfile.backend        # Go 镜像
├── Dockerfile.frontend       # Node 镜像
├── openspec/                 # OpenSpec 规范
│   ├── specs/architecture/  # 项目架构规格
│   └── changes/             # 变更提案
└── .agent/                   # OpenSpec agent skills
```

## 图片存储路径结构
```
{IMAGE_BASE_PATH}/
├── original/landscape/    # 原始横图
├── original/portrait/     # 原始竖图
├── landscape/webp/        # WebP 横图
├── landscape/avif/        # AVIF 横图
├── portrait/webp/         # WebP 竖图
├── portrait/avif/         # AVIF 竖图
└── gif/                   # GIF (不转换)
```

## 错误码体系
- 1000: Internal Server Error
- 1001: Invalid Parameter
- 1002: Unauthorized
- 1003: Forbidden
- 1004: Not Found
- 2000-2003: Image process/upload/delete/list errors
- 3000-3001: Storage/S3 errors

## 关键配置 (.env)
- `API_KEY` — 认证密钥
- `STORAGE_TYPE` — local 或 s3
- `METADATA_STORE_TYPE` — redis
- `REDIS_HOST/PORT/PASSWORD` — Redis 连接
- `S3_ENDPOINT/REGION/BUCKET/ACCESS_KEY/SECRET_KEY` — S3 配置
- `MAX_UPLOAD_COUNT` — 单次上传上限 (默认 20)
- `IMAGE_QUALITY` — WebP/AVIF 质量 1-100 (默认 80)
- `WORKER_THREADS` — libvips 并发线程 (默认 4)
- `WORKER_POOL_SIZE` — 工作池大小 (默认 4)
- `SPEED` — 编码速度 0-8 (默认 5)
- `NEXT_PUBLIC_API_URL` — 前端 API 地址

## 待办
- GitHub fork 仍显示 sync fork (需要 detach)
