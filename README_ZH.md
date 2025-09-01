<div align="center">
  <img src="resource/static/pic/icon_300.webp" width="140"/>
  <h1>GooseForum</h1>
  <p>🚀 现代化的 Go + Vue 3 + TailwindCSS 论坛系统</p>
  
  <p>
    <a href="https://github.com/leancodebox/GooseForum/releases"><img src="https://img.shields.io/github/release/leancodebox/GooseForum.svg" alt="GitHub release"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.23+-blue.svg" alt="Go version"></a>
    <a href="https://vuejs.org"><img src="https://img.shields.io/badge/Vue-3.0+-green.svg" alt="Vue version"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/leancodebox/GooseForum.svg" alt="License"></a>
    <a href="https://github.com/leancodebox/GooseForum/stargazers"><img src="https://img.shields.io/github/stars/leancodebox/GooseForum.svg?style=social" alt="GitHub stars"></a>
  </p>
</div>

## 🌐 Language / 语言

[🇨🇳 中文](README_ZH.md) | [🇺🇸 English](README.md)

## 📖 项目简介

GooseForum 是一个现代化的技术交流社区平台，采用 Go + Vue 3 + TailwindCSS 技术栈开发。提供极简的部署方式和丰富的社区功能，专为技术开发者打造的轻量级论坛系统。

🌐 **在线体验**: [GooseForum](https://gooseforum.online/)

## ✨ 核心特性

### 🎯 用户体系
- **用户注册/登录** - 支持邮箱激活
- **权限管理** - 基于角色的权限控制系统
- **用户中心** - 个人资料管理、头像上传
- **积分系统** - 签到、发帖、回复积分奖励
- **管理后台** - 完整的后台管理功能

### 📝 内容管理
- **文章发布** - 支持 Markdown 编辑器
- **评论系统** - 多级评论回复
- **文章分类** - 灵活的分类管理
- **标签系统** - 文章标签化管理
- **内容审核** - 管理员内容审核功能

### 🛠 技术特性
- **单文件部署** - 编译后单个可执行文件
- **SQLite 支持** - 默认使用 SQLite，支持 MySQL
- **自动备份** - 定时数据库备份
- **响应式设计** - 完美支持移动端
- **主题切换** - 支持明暗主题
- **SEO 友好** - 完整的 SEO 优化

## 🚀 快速开始

### 方式一：下载预编译版本（推荐）

1. 从 [GitHub Releases](https://github.com/leancodebox/GooseForum/releases) 下载对应系统的预编译版本
2. 解压并启动：

```bash
# 解压下载的文件
tar -zxvf GooseForum_Linux_x86_64.tar.gz

# 赋予执行权限
chmod +x ./GooseForum

# 启动服务
./GooseForum serve
```

### 使用 GoReleaser 快速构建

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser@latest

# 构建所有平台
goreleaser build --snapshot --clean

# 构建当前平台
goreleaser build --snapshot --clean --single-target
```



3. 访问 `http://localhost:5234` 开始使用

> 💡 **提示**: 首次启动后，第一个注册的账号将自动成为管理员

### 方式二：从源码构建

#### 环境要求
- Go 1.23+
- Node.js 18+
- npm 或 yarn

#### 构建步骤

```bash
# 克隆项目
git clone https://github.com/leancodebox/GooseForum.git
cd GooseForum

# 构建前端资源
cd resource
npm install
npm run build
cd ..

# 构建后端
go mod tidy
go build -ldflags="-w -s" .

# 启动服务
./GooseForum serve
```

## 🔧 配置说明

GooseForum 启动时会自动创建 `config.toml` 配置文件，主要配置项：

```toml
[server]
port = 5234                    # 服务端口
url = "http://localhost"     # 站点URL

[db.default]
connection = "sqlite"        # 数据库类型 (sqlite/mysql)
path = "./storage/database/sqlite.db"  # SQLite 数据库路径

```

📖 **详细配置说明**: [配置文档](docs/configuration.md)

## 🏗 技术架构

### 后端技术栈
- **Go 1.23+** - 主要开发语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **SQLite/MySQL** - 数据库支持
- **JWT** - 身份认证
- **Viper** - 配置管理
- **Cobra** - 命令行工具

### 前端技术栈
- **Vue 3** - 前端框架 (Composition API)
- **Vite** - 构建工具
- **TailwindCSS 4** - CSS 框架
- **DaisyUI** - UI 组件库
- **TypeScript** - 类型支持
- **Pinia** - 状态管理
- **Vue Router** - 路由管理

### 开发工具
- **Air** - 热重载开发
- **GoReleaser** - 自动化构建发布
- **Vitest** - 前端测试

## 📁 项目结构

```
GooseForum/
├── app/                    # 后端应用代码
│   ├── bundles/           # 工具包
│   ├── console/           # 命令行工具
│   ├── http/              # HTTP 控制器和路由
│   ├── models/            # 数据模型
│   └── service/           # 业务服务
├── resource/              # 前端资源
│   ├── src/               # Vue 源码
│   ├── static/            # 静态资源
│   └── templates/         # Go 模板
├── docs/                  # 项目文档
├── main.go               # 程序入口
└── config.toml           # 配置文件
```

## 🛡 管理功能

### 用户管理
```bash
# 重置管理员密码
./GooseForum user:manage

# 设置用户邮箱
./GooseForum user:set-email
```

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
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request


## 📄 许可证

本项目基于 [MIT License](LICENSE) 开源协议。

## 📚 相关文档

- [配置文档](docs/configuration.md) - 详细的配置选项说明

## 🙏 致谢

感谢所有为 GooseForum 项目做出贡献的开发者！

---

<div align="center">
  <p>如果这个项目对你有帮助，请给我们一个 ⭐️</p>
  <p>Made with ❤️ by <a href="https://github.com/leancodebox">LeanCodeBox</a></p>
</div>
