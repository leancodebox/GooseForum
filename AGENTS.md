# Codex 使用指南（中文）

本文档为 Codex（Codex.ai/code）在本仓库工作时提供指导说明。

## 工程原则

- 代码以可维护为优先：改动保持简洁，命名和结构要清晰；在代码可读性、复用性之间保持平衡，不为了复用而过度抽象，也不重复堆业务逻辑。
- 外部库选择和复杂逻辑实现以稳定、主流、可靠、契合项目为原则；优先使用已经被广泛验证且适合当前技术栈的方案，避免引入小众、维护不稳定或与项目边界不匹配的依赖。

## 项目概览

GooseForum 是一个现代化论坛平台，后端使用 Go，前端和管理后台统一由 `resource/` 下的 Vue 3 应用提供。后端框架采用 Gin，数据库 ORM 使用 GORM，支持 SQLite 与 MySQL。

## 开发命令

### `resource` 构建规则
- 修改 `resource/` 下的模板、样式、前端脚本后，必须根据当前环境重新处理前端产物。
- 当 `config.toml` 中 `[app].env = "local"` 时，使用开发模式：`cd resource && pnpm dev`。
- 当 `config.toml` 中 `[app].env = "production"` 时，使用生产构建：`cd resource && pnpm build`。
- 如果页面样式、资源引用或模板表现异常，优先检查是否遗漏了上面的 `resource` 开发/构建步骤。

### 单独启动服务
```bash
# 后端（热重载）
air

# 前端与管理后台（Vue）
cd resource && pnpm dev
```

### 构建
```bash
# 构建 Go 二进制
go build -ldflags="-w -s" .

# 使用 GoReleaser 构建
goreleaser build --snapshot --clean
goreleaser build --snapshot --clean --single-target  # 仅当前平台

# 构建前端与管理后台（Vue）
cd resource && pnpm build
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
./gooseforum set-user-admin <userId>  # 设置管理员
./gooseforum set-user-email <userId> <email>  # 设置用户邮箱
./gooseforum set-user-password <userId> <password>  # 重置用户密码
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
- 技术栈：Vue 3 + TypeScript + Vite + TailwindCSS 4。
- `src/site/`：主站前端入口、页面与组件。
- `src/admin/`：管理后台入口、布局、页面、组件和独立样式。
- `src/runtime/`：共享 payload runtime 与浏览器辅助逻辑。
- `src/types/`：共享前端类型。
- `templates/`：GoHTML 模板，用于首屏、SEO 与 no-js 降级。

## 数据库

- 默认使用 SQLite，路径：`./storage/database/sqlite.db`
- 支持 MySQL（在 `config.toml` 中配置）
- 首个注册用户自动成为管理员

## 配置

全部配置在 `config.toml` 中，关键项：
- `server.port`：HTTP 端口（默认：5234）
- `db.default.connection`：sqlite 或 mysql
- `app.debug`：是否启用调试模式；不配置时 `local` 默认开启，其他环境默认关闭
