<div align="center">
  <img src="resource/static/pic/icon_300.webp" width="140"/>
  <h1>GooseForum</h1>
  <p>🚀 Modern Go + Vue 3 Forum System</p>

  <p>
    <a href="https://github.com/leancodebox/GooseForum/releases"><img src="https://img.shields.io/github/release/leancodebox/GooseForum.svg" alt="GitHub release"></a>
    <a href="https://pkg.go.dev/github.com/leancodebox/GooseForum"><img src="https://pkg.go.dev/badge/github.com/leancodebox/GooseForum.svg" alt="pkg.go.dev"></a>
    <a href="https://goreportcard.com/report/github.com/leancodebox/GooseForum"><img src="https://goreportcard.com/badge/github.com/leancodebox/GooseForum" alt="Go Report Card"></a>
    <a href="https://app.codecov.io/gh/leancodebox/GooseForum"><img src="https://codecov.io/gh/leancodebox/GooseForum/branch/main/graph/badge.svg" alt="Code coverage"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-blue.svg" alt="Go version"></a>
    <a href="https://tailwindcss.com"><img src="https://img.shields.io/badge/TailwindCSS-4-blue.svg" alt="TailwindCSS"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/leancodebox/GooseForum.svg" alt="License"></a>
    <a href="https://github.com/leancodebox/GooseForum/stargazers"><img src="https://img.shields.io/github/stars/leancodebox/GooseForum.svg?style=social" alt="GitHub stars"></a>
  </p>
</div>

## 🌐 Language / 语言

[🇨🇳 中文](README_ZH.md) | [🇺🇸 English](README.md)

## 📖 Project Overview

GooseForum is a modern technical community platform built with Go + Vue 3 + TailwindCSS. It keeps deployment simple while using a payload-driven SPA experience with no-js HTML rendering for SEO and graceful fallback.

🌐 **Live Demo**: [GooseForum](https://gooseforum.online/)

## ✨ Core Features

### 🎯 User System
- **User Registration/Login** - Email activation support
- **Permission Management** - Role-based access control
- **User Center** - Profile management, avatar upload
- **Points System** - Check-in, posting, reply rewards
- **Admin Panel** - Complete backend management

### 📝 Content Management
- **Article Publishing** - Markdown editor with preview
- **Comment System** - Multi-level replies
- **Article Categories** - Flexible category management
- **Real-time Notifications** - WebSocket-powered notifications
- **Chat System** - Real-time messaging

### 🛠 Technical Features
- **Single File Deployment** - Single executable after compilation
- **SQLite/MySQL Support** - Default SQLite, MySQL optional
- **Auto Backup** - Scheduled database backup
- **Responsive Design** - Perfect mobile support
- **Brand Customization** - Custom logo/text/image support
- **Payload-driven SPA** - Smooth in-app navigation with server-provided page payloads
- **SEO Friendly** - Lightweight no-js GoHTML rendering for crawlers and fallback

## 🚀 Quick Start

### Method 1: Download Pre-compiled Version (Recommended)

1. Download pre-compiled version from [GitHub Releases](https://github.com/leancodebox/GooseForum/releases)
2. Extract and start:

```bash
# Extract
tar -zxvf GooseForum_Linux_x86_64.tar.gz

# Grant permission
chmod +x ./GooseForum

# Start service
./GooseForum serve
```

### Build with GoReleaser

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build all platforms
goreleaser build --snapshot --clean

# Build current platform
goreleaser build --snapshot --clean --single-target
```

3. Visit `http://localhost:5234`

> 💡 **Tip**: First registered account becomes administrator

### Method 2: Build from Source

#### Requirements
- Go 1.26+
- Node.js 18+
- pnpm

#### Build Steps

```bash
# Clone project
git clone https://github.com/leancodebox/GooseForum.git
cd GooseForum

# Build frontend
cd resource && pnpm install && pnpm build && cd ..

# Build backend
go mod tidy
go build -ldflags="-w -s" .

# Start service
./GooseForum serve
```

### Development Mode

```bash
# Backend with hot reload
air

# Public site and admin console frontend
cd resource && pnpm dev
```

The current admin console is served by the `resource` Vue app under `/admin`. It does not require a separate admin frontend service.

## 🔧 Configuration

GooseForum auto-creates `config.toml` on first startup:

```toml
[app]
env = "production"              # local or production
# debug is optional; local defaults to true, other environments default to false

[server]
port = 5234                    # Service port
url = "http://localhost"       # Site URL

[db.default]
connection = "sqlite"          # Database type (sqlite/mysql)
path = "./storage/database/sqlite.db"
```

📖 **Detailed Configuration**: [Configuration Documentation](docs/user/configuration.md)

## 🏗 Technical Architecture

### Backend Tech Stack
- **Go 1.26+** - Main language
- **Gin** - Web framework
- **GORM** - ORM
- **SQLite/MySQL** - Database
- **JWT** - Authentication
- **Cobra** - CLI

### Frontend Tech Stack
- **Vue 3** - Public site and admin UI framework
- **TypeScript** - Type-safe frontend code
- **Payload SPA Runtime** - Client-side navigation via `X-Goose-Page` JSON payloads
- **TailwindCSS 4** - CSS framework
- **GoHTML** - Lightweight no-js/SEO templates
- **Vite** - Build tool

### Admin Panel Tech Stack
- **Vue 3 + TypeScript** - Admin UI inside `resource/src/admin`
- **TailwindCSS 4** - Admin-scoped styling and design tokens
- **Reka UI / VueUse** - Accessible primitives and interaction utilities where needed
- **Unovis** - Admin charts and statistics visualization
- **SortableJS / vuedraggable** - Drag sorting for operational lists

## 📁 Project Structure

```
GooseForum/
├── app/                    # Backend code
│   ├── bundles/           # Utilities (JWT, cache, events)
│   ├── console/           # CLI commands
│   ├── http/              # Controllers, middleware, routes
│   ├── models/            # GORM models
│   └── service/           # Business services
├── resource/              # Frontend resources
│   ├── src/
│   │   ├── site/          # Public Vue app
│   │   ├── admin/         # Admin Vue app
│   │   ├── runtime/       # Shared payload runtime
│   │   ├── styles/        # Public-site styles
│   │   └── types/         # Shared frontend types
│   ├── static/            # Static assets
│   └── templates/         # No-js/SEO GoHTML templates
├── docs/                  # Documentation
├── main.go               # Entry point
└── config.toml           # Configuration
```

## 🛡 Admin Features

```bash
# Grant administrator role
./GooseForum set-user-admin <userId>

# Set user email
./GooseForum set-user-email <userId> <email>

# Set user password
./GooseForum set-user-password <userId> <password>
```

### Admin Panel Features
- **User Management** - Search, filter, ban, delete users
- **Site Settings** - General, brand, footer, mail, security, posting
- **Category Management** - Create, edit, delete categories
- **Sponsorship Management** - Sponsor tiers and user sponsors
- **Dashboard** - Traffic stats, daily analytics

### Data Backup
- Automatic scheduled SQLite backup
- Configurable frequency and retention
- Backup in `./storage/databasebackup/`

## 📦 Deployment

### Production
1. Use reverse proxy (Nginx/Apache)
2. Configure HTTPS
3. Set up scheduled backups
4. Monitor logs

### Docker
```dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY GooseForum .
CMD ["./GooseForum", "serve"]
```

## 🤝 Contributing

1. Fork this project
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push branch (`git push origin feature/AmazingFeature`)
5. Create Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE)

## 📚 Related Documentation

- [Documentation Index](docs/README.md)
- [Configuration Documentation](docs/user/configuration.md)
- [Resource Frontend Design](docs/architecture/resource-frontend.md)
- [Resource UI Specification](docs/frontend/ui-spec.md)
- [Article Pinning Design](docs/features/article-pinning.md)
- [Chinese README](README_ZH.md)

## 🙏 Acknowledgments

Thanks to all contributors!

---

<div align="center">
  <p>If this project helps you, please give us a ⭐️</p>
  <p>Made with ❤️ by <a href="https://github.com/leancodebox">LeanCodeBox</a></p>
</div>
