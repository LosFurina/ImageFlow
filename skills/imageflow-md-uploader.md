---
name: imageflow-md-uploader
title: ImageFlow Markdown 图片上传工具
description: AI Agent 可调用的 CLI 工具，自动扫描 Markdown 中的本地图片引用、上传至 ImageFlow 图床并替换为远程 URL。独立于 ImageFlow 项目本身，通过 OpenAPI 调用。
version: 1.0.0
tags: [imageflow, markdown, cli, agent-tool, image-upload]
---

# ImageFlow Markdown 图片上传工具

## 工具定位

一个独立的 Go CLI 工具，不依赖 ImageFlow 项目本体代码。它通过 ImageFlow 的 OpenAPI 上传端点将 Markdown 中的本地图片自动上传到图床，并将图片引用替换为可访问的远程 URL。

**典型使用场景**：
- AI Agent 优化博客文章时，文章中引用了本地图片，需要先上传到图床才能继续处理
- 批量迁移本地 Markdown 文档中的图片到远程图床
- 静态站点生成时的图片资源自动上传

## 安装

### 方式一：直接从源码构建（推荐）

```bash
git clone https://github.com/LosFurina/ImageFlow.git
cd ImageFlow
go build -o md-uploader ./cmd/md-uploader/
# 可选：移动到 PATH
mv md-uploader /usr/local/bin/
```

### 方式二：go install（如果项目已发布 module）

```bash
go install github.com/LosFurina/ImageFlow/cmd/md-uploader@latest
```

## 环境变量配置

**必须通过环境变量传入凭证，禁止命令行 flag**（避免泄露到 shell history / ps / 日志）。

```bash
export IMAGEFLOW_AK=<your_access_key>
export IMAGEFLOW_SK=<your_secret_key>
```

**可选配置**：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `IMAGEFLOW_AK` | Access Key | 必填 |
| `IMAGEFLOW_SK` | Secret Key | 必填 |
| `IMAGEFLOW_HOST` | ImageFlow 后端地址 | `http://localhost:8686` |

## 使用方式

### 基本命令

```bash
# 输出到 stdout（不修改原文件）
md-uploader /path/to/article.md

# 原地修改（推荐用于 Agent 自动化）
md-uploader -i /path/to/article.md

# 输出到新文件
md-uploader -o /path/to/output.md /path/to/article.md
```

### 输入输出示例

**输入** (`article.md`):
```markdown
# 我的文章

这是一张本地图片：
![screenshot](./images/demo.png)

这是一张已经在线的图片（不会被处理）：
![logo](https://example.com/logo.png)
```

**输出**:
```markdown
# 我的文章

这是一张本地图片：
![screenshot](https://your-imageflow-host/api/images/abc123.png)

这是一张已经在线的图片（不会被处理）：
![logo](https://example.com/logo.png)
```

## AI Agent 集成指南

### 集成步骤

1. **确认环境变量已设置**
   - 在启动 Agent 的环境中提前注入 `IMAGEFLOW_AK` 和 `IMAGEFLOW_SK`
   - 或在 Agent 的系统提示词中说明需要设置这两个变量

2. **检测本地图片**
   - 读取 Markdown 文件内容
   - 检查是否包含本地图片引用（匹配 `![alt](local/path.png)` 且路径不以 `http` 开头）

3. **上传处理**
   - 如果存在本地图片，调用：
     ```bash
     md-uploader -i /path/to/article.md
     ```
   - 工具会自动计算 AK/SK 签名、上传图片、替换 URL

4. **继续优化**
   - 处理完成后，文章中的图片引用已是可公开访问的 URL
   - Agent 可继续进行文章优化、格式调整等操作

### Agent 系统提示词模板

如果你是为 Agent 配置系统提示词，可以加入以下内容：

```
当处理 Markdown 文档时，如果发现文档中引用了本地图片（如 ![alt](./images/xxx.png)），需要先将图片上传到图床。

使用 imageflow-md-uploader 工具：
1. 确保环境变量 IMAGEFLOW_AK 和 IMAGEFLOW_SK 已设置
2. 执行：md-uploader -i <markdown_file_path>
3. 等待上传完成，文件中的本地图片引用会被自动替换为图床 URL
4. 上传完成后继续处理文章

注意：AK/SK 只能通过环境变量传入，不要使用命令行参数。
```

## 技术细节

### 认证机制

- 采用 HMAC-SHA256 签名认证
- 签名字符串格式：`Method\nPath\nTimestamp\nSHA256(Body)`
- HTTP Headers: `X-Access-Key`, `X-Signature`, `X-Timestamp`

### 支持的图片引用格式

- 标准 Markdown: `![alt text](path/to/image.png)`
- 相对路径: `./images/photo.jpg`, `../assets/chart.png`
- 绝对路径: `/Users/name/blog/images/pic.webp`
- **自动排除**以 `http://` 或 `https://` 开头的已在线图片

### OpenAPI 端点

```
POST /openapi/upload
Content-Type: multipart/form-data

Headers:
  X-Access-Key: <ak>
  X-Signature: <hmac_signature>
  X-Timestamp: <unix_timestamp>

Body:
  images: <binary_file_data>

Response:
  {
    "images": [
      {"url": "https://host/api/images/xxx.jpg", "filename": "xxx.jpg"}
    ]
  }
```

## 常见问题

### 上传失败

1. 检查 `IMAGEFLOW_AK`/`IMAGEFLOW_SK` 是否正确
2. 确认 ImageFlow 服务可达（默认 `http://localhost:8686`）
3. 检查图片文件是否存在且有读取权限

### 路径解析错误

- 确保 Markdown 文件路径与图片路径的相对关系正确
- 工具会基于 Markdown 文件所在目录解析相对路径

### 编译失败

```bash
# 确保在项目根目录执行
cd ImageFlow/
go mod tidy
go build ./cmd/md-uploader/
```

## 安全规范

1. **AK/SK 传递方式**
   - ✓ 环境变量: `export IMAGEFLOW_AK=xxx`
   - ✗ 命令行参数: ~~`md-uploader -ak xxx -sk xxx`~~ 已移除

2. **原因**
   - 命令行参数会进入 shell history
   - `ps` 命令可以看到进程参数
   - Agent 会话日志可能记录命令行

3. **最佳实践**
   - 在 `.bashrc`/`.zshrc` 或启动脚本中设置环境变量
   - 使用 Docker secrets 或密钥管理工具注入
   - CI/CD 中使用 GitHub Actions secrets
