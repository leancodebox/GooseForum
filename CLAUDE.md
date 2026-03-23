# CLAUDE 使用指南（中文）

本文档为 Claude Code（claude.ai/code）在本仓库工作时提供指导说明。

## 项目概览

GooseForum 是一个现代化论坛平台，后端使用 Go，前端包含 Vue 3 与 React 管理端。后端框架采用 Gin，数据库 ORM 使用 GORM，支持 SQLite 与 MySQL。

## 开发命令

### 一体化开发环境
```bash
./dev.sh
```
同时启动三个服务：
- 后端（Go）：http://localhost:5234
- 前端（Vue）：http://localhost:3009
- 管理端（React）：http://localhost:5173

### 单独启动服务
```bash
# 后端（热重载）
air

# 前端（Vue）
cd resource && pnpm dev

# 管理端（React）
cd admin && pnpm dev
```

### 构建
```bash
# 构建 Go 二进制
go build -ldflags="-w -s" .

# 使用 GoReleaser 构建
goreleaser build --snapshot --clean
goreleaser build --snapshot --clean --single-target  # 仅当前平台

# 构建前端（Vue）
cd resource && pnpm build

# 构建管理端（React）
cd admin && pnpm build
```

### 测试
```bash
# 后端 Go 测试
go test ./...

# 运行指定测试
go test -v ./path/to/package -run TestName

# 前端测试（Vitest）
cd resource && npx vitest run
```

### CLI 命令
```bash
./gooseforum serve                 # 启动服务
./gooseforum user:manage           # 重置管理员密码
./gooseforum user:set-email <email>  # 设置用户邮箱
```

## 架构

### 后端目录结构（`app/`）
- `bundles/`：通用工具包（JWT、配置、分页、缓存等）
- `console/cmd/`：CLI 命令（基于 Cobra）
- `http/controllers/`：HTTP 控制器
- `http/middleware/`：Gin 中间件（CORS、认证等）
- `http/routes/`：路由定义
- `models/`：GORM 模型（用户、文章、回复等）
- `service/`：业务服务

### 前端目录结构（`resource/`）
- 技术栈：Vue 3 + TypeScript + Vite
- 使用 TailwindCSS 4、PrimeVue、Pinia、Vue Router
- 组件位于 `src/components/`，页面位于 `src/views/`

### 管理端（`admin/`）
- 技术栈：React 19 + TypeScript + Vite
- 使用 TailwindCSS 4、TanStack Query、TanStack Router、Radix UI

## 数据库

- 默认使用 SQLite，路径：`./storage/database/sqlite.db`
- 支持 MySQL（在 `config.toml` 中配置）
- 首个注册用户自动成为管理员

## 配置

全部配置在 `config.toml` 中，关键项：
- `server.port`：HTTP 端口（默认：5234）
- `db.default.connection`：sqlite 或 mysql
- `app.debug`：是否启用调试模式
