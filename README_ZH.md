<div align="center">
  <img src="resource/static/pic/icon_300.webp" width="140"/>
  <h1>GooseForum</h1>
  <p>🚀 现代化的 Go + Alpine.js 论坛系统</p>

  <p>
    <a href="https://github.com/leancodebox/GooseForum/releases"><img src="https://img.shields.io/github/release/leancodebox/GooseForum.svg" alt="GitHub release"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-blue.svg" alt="Go version"></a>
    <a href="https://tailwindcss.com"><img src="https://img.shields.io/badge/TailwindCSS-4-blue.svg" alt="TailwindCSS"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/leancodebox/GooseForum.svg" alt="License"></a>
    <a href="https://github.com/leancodebox/GooseForum/stargazers"><img src="https://img.shields.io/github/stars/leancodebox/GooseForum.svg?style=social" alt="GitHub stars"></a>
  </p>
</div>

## 🌐 Language / 语言

[🇨🇳 中文](README_ZH.md) | [🇺🇸 English](README.md)

## 📖 项目简介

GooseForum 是一个现代化的技术交流社区平台，采用 Go + Alpine.js + TailwindCSS 技术栈开发。提供极简的部署方式和丰富的社区功能，专为技术开发者打造的轻量级论坛系统。

🌐 **在线体验**: [GooseForum](https://gooseforum.online/)

## ✨ 核心特性

### 🎯 用户体系
- **用户注册/登录** - 支持邮箱激活
- **权限管理** - 基于角色的权限控制
- **用户中心** - 个人资料管理、头像上传
- **积分系统** - 签到、发帖、回复奖励
- **管理后台** - 完整的后台管理功能

### 📝 内容管理
- **文章发布** - Markdown 编辑器+预览
- **评论系统** - 多级评论回复
- **文章分类** - 灵活的分类管理
- **实时通知** - WebSocket 驱动通知
- **聊天系统** - 实时消息

### 🛠 技术特性
- **单文件部署** - 编译后单个可执行文件
- **SQLite/MySQL 支持** - 默认 SQLite，可选 MySQL
- **自动备份** - 定时数据库备份
- **响应式设计** - 完美支持移动端
- **品牌定制** - 支持自定义 Logo/文字/图片
- **SEO 友好** - 完整的 SEO 优化

## 🚀 快速开始

### 方式一：下载预编译版本（推荐）

1. 从 [GitHub Releases](https://github.com/leancodebox/GooseForum/releases) 下载预编译版本
2. 解压并启动：

```bash
# 解压
tar -zxvf GooseForum_Linux_x86_64.tar.gz

# 赋予权限
chmod +x ./GooseForum

# 启动服务
./GooseForum serve
```

### 使用 GoReleaser 构建

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser@latest

# 构建所有平台
goreleaser build --snapshot --clean

# 构建当前平台
goreleaser build --snapshot --clean --single-target
```

3. 访问 `http://localhost:5234`

> 💡 **提示**: 首次启动后，第一个注册的账号将自动成为管理员

### 方式二：从源码构建

#### 环境要求
- Go 1.24+
- Node.js 18+
- npm 或 pnpm

#### 构建步骤

```bash
# 克隆项目
git clone https://github.com/leancodebox/GooseForum.git
cd GooseForum

# 构建前端
cd resource && pnpm install && pnpm build && cd ..

# 构建后端
go mod tidy
go build -ldflags="-w -s" .

# 启动服务
./GooseForum serve
```

### 开发模式

```bash
# 启动所有服务（后端 + 前端 + 管理后台）
./dev.sh

# 或单独运行：
air                         # 后端热重载
cd resource && pnpm dev      # Vue 前端
cd admin && pnpm dev         # React 管理后台
```

## 🔧 配置说明

GooseForum 启动时自动创建 `config.toml`：

```toml
[server]
port = 5234                    # 服务端口
url = "http://localhost"         # 站点 URL

[db.default]
connection = "sqlite"            # 数据库类型 (sqlite/mysql)
path = "./storage/database/sqlite.db"
```

📖 **详细配置说明**: [配置文档](docs/configuration.md)

## 🏗 技术架构

### 后端技术栈
- **Go 1.24+** - 主要开发语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **SQLite/MySQL** - 数据库支持
- **JWT** - 身份认证
- **Cobra** - 命令行工具

### 前端技术栈
- **Alpine.js** - 轻量级 JS 框架
- **TailwindCSS 4** - CSS 框架
- **GoHTML** - 服务器端模板
- **Vite** - 构建工具

### 管理后台技术栈
- **React 19** - UI 框架
- **TypeScript** - 类型支持
- **shadcn-admin** - 管理后台模板
- **TanStack Query/Router** - 数据获取和路由
- **Radix UI** - 组件库

## 📁 项目结构

```
GooseForum/
├── app/                    # 后端代码
│   ├── bundles/           # 工具包（JWT、缓存、事件）
│   ├── console/           # CLI 命令
│   ├── http/              # 控制器、中间件、路由
│   ├── models/            # GORM 模型
│   └── service/           # 业务服务
├── resource/              # 前端资源
│   ├── src/               # Alpine.js 源码
│   ├── static/            # 静态资源
│   └── templates/         # GoHTML 模板
├── admin/                 # React 管理后台
├── docs/                  # 文档
├── main.go               # 程序入口
└── config.toml           # 配置文件
```

## 🛡 管理功能

```bash
# 重置管理员密码
./GooseForum user:manage

# 设置用户邮箱
./GooseForum user:set-email <邮箱>
```

### 管理后台功能
- **用户管理** - 搜索、筛选、封禁、删除用户
- **站点设置** - 基本信息、品牌设置、Footer、邮件、安全、发帖设置
- **分类管理** - 创建、编辑、删除分类
- **赞助商管理** - 赞助商等级和用户赞助记录
- **仪表盘** - 流量统计、每日数据

### 数据备份
- 自动定时备份 SQLite 数据库
- 可配置备份频率和保留数量
- 备份文件存储在 `./storage/databasebackup/` 目录

## 🔄 开发模式

```bash
# 安装 Air 热重载工具
go install github.com/cosmtrek/air@latest

# 启动开发模式
air

# 前端开发模式
cd resource
npm run dev
```

## 📦 部署建议

### 生产环境部署
1. 使用反向代理 (Nginx/Apache)
2. 配置 HTTPS 证书
3. 设置定时备份
4. 监控日志文件

### Docker 部署
```dockerfile
# Dockerfile 示例
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY GooseForum .
CMD ["./GooseForum", "serve"]
```

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源协议。

## 📚 相关文档

- [配置文档](docs/configuration.md) - 详细的配置选项说明
- [English README](README.md)

## 🙏 致谢

感谢所有为 GooseForum 项目做出贡献的开发者！

---

<div align="center">
  <p>如果这个项目对你有帮助，请给我们一个 ⭐️</p>
  <p>Made with ❤️ by <a href="https://github.com/leancodebox">LeanCodeBox</a></p>
</div>