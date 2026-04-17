<div align="center">
  <img src="resource/static/pic/icon_300.webp" width="140"/>
  <h1>GooseForum</h1>
  <p>🚀 Modern Go + Alpine.js Forum System</p>

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

## 📖 Project Overview

GooseForum is a modern technical community platform built with Go + Alpine.js + TailwindCSS. It provides simple deployment and rich community features, designed as a lightweight forum system for technical developers.

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
- **SEO Friendly** - Complete SEO optimization

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
- Go 1.24+
- Node.js 18+
- npm or pnpm

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
# Start all services (backend + frontend + admin)
./dev.sh

# Or run individually:
air                    # Backend with hot reload
cd resource && pnpm dev    # Vue frontend
cd admin && pnpm dev      # React admin panel
```

## 🔧 Configuration

GooseForum auto-creates `config.toml` on first startup:

```toml
[server]
port = 5234                    # Service port
url = "http://localhost"       # Site URL

[db.default]
connection = "sqlite"          # Database type (sqlite/mysql)
path = "./storage/database/sqlite.db"
```

📖 **Detailed Configuration**: [Configuration Documentation](docs/configuration.md)

## 🏗 Technical Architecture

### Backend Tech Stack
- **Go 1.24+** - Main language
- **Gin** - Web framework
- **GORM** - ORM
- **SQLite/MySQL** - Database
- **JWT** - Authentication
- **Cobra** - CLI

### Frontend Tech Stack
- **Alpine.js** - Lightweight JS framework
- **TailwindCSS 4** - CSS framework
- **GoHTML** - Server-side templates
- **Vite** - Build tool

### Admin Panel Tech Stack
- **React 19** - UI framework
- **TypeScript** - Type safety
- **shadcn-admin** - Admin template
- **TanStack Query/Router** - Data fetching & routing
- **Radix UI** - Component library

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
│   ├── src/               # Alpine.js source
│   ├── static/            # Static assets
│   └── templates/         # GoHTML templates
├── admin/                 # React admin panel
├── docs/                  # Documentation
├── main.go               # Entry point
└── config.toml           # Configuration
```

## 🛡 Admin Features

```bash
# Reset admin password
./GooseForum user:manage

# Set user email
./GooseForum user:set-email <email>
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

- [Configuration Documentation](docs/configuration.md)
- [Chinese README](README_ZH.md)

## 🙏 Acknowledgments

Thanks to all contributors!

---

<div align="center">
  <p>If this project helps you, please give us a ⭐️</p>
  <p>Made with ❤️ by <a href="https://github.com/leancodebox">LeanCodeBox</a></p>
</div>