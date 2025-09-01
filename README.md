<div align="center">
  <img src="resource/static/pic/icon_300.webp" width="140"/>
  <h1>GooseForum</h1>
  <p>🚀 Modern Go + Vue 3 + TailwindCSS Forum System</p>
  
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

## 📖 Project Overview

GooseForum is a modern technical community platform built with Go + Vue 3 + TailwindCSS technology stack. It provides simple deployment and rich community features, designed as a lightweight forum system specifically for technical developers.

🌐 **Live Demo**: [GooseForum](https://gooseforum.online/)

## ✨ Core Features

### 🎯 User System
- **User Registration/Login** - Email activation support
- **Permission Management** - Role-based access control system
- **User Center** - Profile management, avatar upload
- **Points System** - Check-in, posting, reply point rewards
- **Admin Panel** - Complete backend management functionality

### 📝 Content Management
- **Article Publishing** - Markdown editor support
- **Comment System** - Multi-level comment replies
- **Article Categories** - Flexible category management
- **Tag System** - Article tagging management
- **Content Moderation** - Admin content review functionality

### 🛠 Technical Features
- **Single File Deployment** - Single executable file after compilation
- **SQLite Support** - Default SQLite, MySQL support
- **Auto Backup** - Scheduled database backup
- **Responsive Design** - Perfect mobile support
- **Theme Switching** - Light/dark theme support
- **SEO Friendly** - Complete SEO optimization

## 🚀 Quick Start

### Method 1: Download Pre-compiled Version (Recommended)

1. Download the pre-compiled version for your system from [GitHub Releases](https://github.com/leancodebox/GooseForum/releases)
2. Extract and start:

```bash
# Extract the downloaded file
tar -zxvf GooseForum_Linux_x86_64.tar.gz

# Grant execute permission
chmod +x ./GooseForum

# Start service
./GooseForum serve
```

### Quick Build with GoReleaser

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build for all platforms
goreleaser build --snapshot --clean

# Build for current platform
goreleaser build --snapshot --clean --single-target
```

3. Visit `http://localhost:5234` to start using

> 💡 **Tip**: After first startup, the first registered account will automatically become an administrator

### Method 2: Build from Source

#### Requirements
- Go 1.23+
- Node.js 18+
- npm or yarn

#### Build Steps

```bash
# Clone project
git clone https://github.com/leancodebox/GooseForum.git
cd GooseForum

# Build frontend resources
cd resource
npm install
npm run build
cd ..

# Build backend
go mod tidy
go build -ldflags="-w -s" .

# Start service
./GooseForum serve
```

## 🔧 Configuration

GooseForum will automatically create a `config.toml` configuration file on startup. Main configuration items:

```toml
[server]
port = 5234                    # Service port
url = "http://localhost"     # Site URL

[db.default]
connection = "sqlite"        # Database type (sqlite/mysql)
path = "./storage/database/sqlite.db"  # SQLite database path
```

📖 **Detailed Configuration**: [Configuration Documentation](docs/configuration.md)

## 🏗 Technical Architecture

### Backend Tech Stack
- **Go 1.23+** - Main development language
- **Gin** - Web framework
- **GORM** - ORM framework
- **SQLite/MySQL** - Database support
- **JWT** - Authentication
- **Viper** - Configuration management
- **Cobra** - Command line tool

### Frontend Tech Stack
- **Vue 3** - Frontend framework (Composition API)
- **Vite** - Build tool
- **TailwindCSS 4** - CSS framework
- **DaisyUI** - UI component library
- **TypeScript** - Type support
- **Pinia** - State management
- **Vue Router** - Route management

### Development Tools
- **Air** - Hot reload development
- **GoReleaser** - Automated build and release
- **Vitest** - Frontend testing

## 📁 Project Structure

```
GooseForum/
├── app/                    # Backend application code
│   ├── bundles/           # Utility packages
│   ├── console/           # Command line tools
│   ├── http/              # HTTP controllers and routes
│   ├── models/            # Data models
│   └── service/           # Business services
├── resource/              # Frontend resources
│   ├── src/               # Vue source code
│   ├── static/            # Static assets
│   └── templates/         # Go templates
├── docs/                  # Project documentation
├── main.go               # Program entry point
└── config.toml           # Configuration file
```

## 🛡 Admin Features

### User Management
```bash
# Reset admin password
./GooseForum user:manage

# Set user email
./GooseForum user:set-email
```

### Data Backup
- Automatic scheduled SQLite database backup
- Configurable backup frequency and retention count
- Backup files stored in `./storage/databasebackup/` directory

## 🔄 Development Mode

```bash
# Install Air hot reload tool
go install github.com/cosmtrek/air@latest

# Start development mode
air

# Frontend development mode
cd resource
npm run dev
```

## 📦 Deployment Recommendations

### Production Environment Deployment
1. Use reverse proxy (Nginx/Apache)
2. Configure HTTPS certificates
3. Set up scheduled backups
4. Monitor log files

### Docker Deployment
```dockerfile
# Dockerfile example
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY GooseForum .
CMD ["./GooseForum", "serve"]
```

## 🤝 Contributing

1. Fork this project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Create a Pull Request

## 📄 License

This project is open source under the [MIT License](LICENSE).

## 📚 Related Documentation

- [Configuration Documentation](docs/configuration.md) - Detailed configuration options

## 🙏 Acknowledgments

Thanks to all developers who have contributed to the GooseForum project!

---

<div align="center">
  <p>If this project helps you, please give us a ⭐️</p>
  <p>Made with ❤️ by <a href="https://github.com/leancodebox">LeanCodeBox</a></p>
</div>