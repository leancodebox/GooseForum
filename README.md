<div align="center">
  <img src="resource/static/pic/icon_300.webp" width="140"/>
  <h1>GooseForum</h1>
  <p>🚀 Modern Go + Vue 3 Forum System</p>

  <p>
    <a href="https://github.com/leancodebox/GooseForum/releases"><img src="https://img.shields.io/github/release/leancodebox/GooseForum.svg" alt="GitHub release"></a>
    <a href="https://pkg.go.dev/github.com/leancodebox/GooseForum"><img src="https://pkg.go.dev/badge/github.com/leancodebox/GooseForum.svg" alt="pkg.go.dev"></a>
    <a href="https://goreportcard.com/report/github.com/leancodebox/GooseForum"><img src="https://goreportcard.com/badge/github.com/leancodebox/GooseForum" alt="Go Report Card"></a>
    <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-blue.svg" alt="Go version"></a>
    <a href="https://tailwindcss.com"><img src="https://img.shields.io/badge/TailwindCSS-4-blue.svg" alt="TailwindCSS"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/leancodebox/GooseForum.svg" alt="License"></a>
    <a href="https://github.com/leancodebox/GooseForum/stargazers"><img src="https://img.shields.io/github/stars/leancodebox/GooseForum.svg?style=social" alt="GitHub stars"></a>
  </p>

  <p><a href="README_ZH.md">中文</a> | <a href="README.md">English</a></p>
</div>

![GooseForum interface preview](https://github.com/leancodebox/assert/blob/main/gooseforum-readme-poster.png?raw=true)

## Quick Start

### Download and Run

Download the latest prebuilt binary from [GitHub Releases](https://github.com/leancodebox/GooseForum/releases), then start it:

```bash
tar -zxvf GooseForum_Linux_x86_64.tar.gz
chmod +x ./GooseForum
./GooseForum serve
```

Open `http://localhost:5234`. The first registered user automatically becomes the administrator.

### Build from Source

Requirements:

- Go 1.26+
- Node.js 18+
- pnpm

```bash
git clone https://github.com/leancodebox/GooseForum.git
cd GooseForum

cd resource && pnpm install && pnpm build && cd ..
go mod tidy
go build -ldflags="-w -s" .

./GooseForum serve
```

### Configuration

GooseForum creates `config.toml` on first startup. The default database is SQLite.

```toml
[app]
env = "production"

[server]
port = 5234
url = "http://localhost"

[db.default]
connection = "sqlite"
path = "./storage/database/sqlite.db"
```

See [configuration documentation](docs/user/configuration.md) for MySQL, mail, backup, security, and site settings.

### Admin Commands

```bash
./GooseForum set-user-admin <userId>
./GooseForum set-user-email <userId> <email>
./GooseForum set-user-password <userId> <password>
```

## What Is GooseForum?

GooseForum is a technical community platform built with Go, Gin, GORM, Vue 3, TypeScript, Vite, and TailwindCSS. It ships as a single executable, supports SQLite/MySQL, and provides a payload-driven SPA experience with server-rendered fallback pages for SEO and no-js access.

Live demo: [gooseforum.online](https://gooseforum.online/)

## Features

- Markdown topics, replies, categories, notifications, chat, drafts, and user profiles.
- Role and permission management with a full admin console.
- Responsive public UI for desktop and mobile.
- Theme workbench for light/dark theme preview and publishing.
- SQLite by default, optional MySQL, scheduled backups.
- Payload-driven navigation with no-js GoHTML templates.
- Brand customization for logo, text, footer, and site assets.

## Development

```bash
# Backend with hot reload
air

# Public site and admin console
cd resource && pnpm dev
```

The admin console is served by the same Vue app under `/admin`; it does not require a separate frontend service.

## Project Structure

```text
GooseForum/
├── app/                    # Backend code
│   ├── console/            # CLI commands
│   ├── http/               # Controllers, middleware, routes
│   ├── models/             # GORM models
│   └── service/            # Business services
├── resource/               # Vue 3 frontend, templates, static assets
│   ├── src/site/           # Public site
│   ├── src/admin/          # Admin console
│   ├── src/runtime/        # Payload runtime and shared browser helpers
│   └── templates/          # GoHTML fallback templates
├── docs/                   # Documentation
├── main.go
└── config.toml
```

## Deployment Notes

For production, place GooseForum behind a reverse proxy such as Nginx or Caddy, enable HTTPS, and configure database backups.

Minimal container image:

```dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY GooseForum .
CMD ["./GooseForum", "serve"]
```

## Documentation

- [Documentation Index](docs/README.md)
- [Configuration](docs/user/configuration.md)
- [Frontend Architecture](docs/architecture/resource-frontend.md)
- [UI Specification](docs/frontend/ui-spec.md)
- [中文 README](README_ZH.md)

## License

MIT License. See [LICENSE](LICENSE).
