---
name: imageflow-ops
title: ImageFlow 项目运维操作
description: ImageFlow 图床服务的 Docker Compose 部署、日常运维、md-uploader CLI 使用流程
version: 1.0.0
tags: [imageflow, docker-compose, ghcr, markdown]
---

# ImageFlow 项目运维操作

## 前置信息

- **项目路径**: `~/Desktop/workspace/repos/ImageFlow-LosFurina`
- **GitHub**: `LosFurina/ImageFlow`
- **GHCR 镜像**:
  - `ghcr.io/losfurina/imageflow-backend:latest`
  - `ghcr.io/losfurina/imageflow-frontend:latest`
- **端口**: 后端 8686, 前端 3000
- **AK/SK**: 通过环境变量 `IMAGEFLOW_AK` / `IMAGEFLOW_SK` 注入

## 1. 启动服务（生产模式）

从 GHCR 拉取预构建镜像，适合服务器部署或本地快速启动：

```bash
cd ~/Desktop/workspace/repos/ImageFlow-LosFurina
docker compose up -d
```

首次会拉取镜像，后续 `up -d` 会直接使用本地缓存的镜像。需要强制更新时：

```bash
docker compose pull && docker compose up -d
```

## 2. 启动服务（开发模式）

本地修改代码后重新构建：

```bash
cd ~/Desktop/workspace/repos/ImageFlow-LosFurina
docker compose -f docker-compose.dev.yaml up -d --build
```

## 3. 停止服务

```bash
cd ~/Desktop/workspace/repos/ImageFlow-LosFurina
docker compose down
# 或开发模式
docker compose -f docker-compose.dev.yaml down
```

## 4. 查看日志

```bash
# 全部服务
docker compose logs -f

# 仅后端
docker compose logs -f backend

# 仅前端
docker compose logs -f frontend
```

## 5. 更新镜像（手动触发 CI）

GitHub Actions workflow 文件: `.github/workflows/docker-build.yml`

```bash
# 推送 main 分支会自动触发
# 或手动触发 via gh CLI
cd ~/Desktop/workspace/repos/ImageFlow-LosFurina
git push origin main
```

CI 构建完成后，拉取最新镜像：

```bash
docker compose pull && docker compose up -d
```

## 6. md-uploader CLI 使用

**编译**:

```bash
cd ~/Desktop/workspace/repos/ImageFlow-LosFurina
go build -o md-uploader ./cmd/md-uploader/
```

**使用前设置 AK/SK**（必须环境变量，禁止命令行传参）:

```bash
export IMAGEFLOW_AK=<your_access_key>
export IMAGEFLOW_SK=<your_secret_key>
```

**处理单篇 Markdown**:

```bash
# 输出到 stdout
./md-uploader article.md > article-updated.md

# 输出到新文件
./md-uploader -o article-updated.md article.md

# 原地修改
./md-uploader -i article.md
```

**AI Agent 工作流集成**:

Agent 在优化 Markdown 文章时，若发现本地图片引用，执行：

```bash
cd ~/Desktop/workspace/repos/ImageFlow-LosFurina
export IMAGEFLOW_AK=... && export IMAGEFLOW_SK=...
go run ./cmd/md-uploader/main.go -i /path/to/article.md
```

上传完成后，图片引用自动替换为 ImageFlow 图床 URL。

## 7. 验证服务状态

```bash
# 后端健康检查
curl http://localhost:8686/health

# 前端访问
open http://localhost:3000
```

## 8. 常见问题

### 8.1 端口冲突

若 8686 或 3000 被占用，修改 `docker-compose.yaml` 中的 `ports` 映射。

### 8.2 GHCR 镜像拉取失败

确认镜像为 public：

```bash
gh api users/LosFurina/packages/container/imageflow-backend | jq '.visibility'
```

### 8.3 md-uploader 编译失败

确保 Go 版本 >= 1.21，且在项目根目录执行：

```bash
go mod tidy
go build ./cmd/md-uploader/
```

## 9. 安全红线

- **AK/SK 绝对禁止通过命令行 flag 传入**（`-ak` / `-sk` 已被移除）
- 仅允许通过 `IMAGEFLOW_AK` 和 `IMAGEFLOW_SK` 环境变量注入
- 避免泄露到 shell history、`ps` 输出、Agent 日志
