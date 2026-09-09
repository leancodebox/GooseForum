<div align="center">
  <img src="resource/static/pic/icon_300.webp?t=1788958424" width="140"/>
  <h1>GooseForum</h1>
  <p>A forum for conversations, shared knowledge, and lasting communities.</p>

  <p>
    <a href="https://github.com/leancodebox/GooseForum/releases"><img src="https://img.shields.io/github/release/leancodebox/GooseForum.svg" alt="GitHub release"></a>
    <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-blue.svg" alt="Go version"></a>
    <a href="https://tailwindcss.com"><img src="https://img.shields.io/badge/TailwindCSS-4-blue.svg" alt="TailwindCSS"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/leancodebox/GooseForum.svg" alt="License"></a>
    <a href="https://github.com/leancodebox/GooseForum/stargazers"><img src="https://img.shields.io/github/stars/leancodebox/GooseForum.svg?style=social" alt="GitHub stars"></a>
  </p>

  <p><a href="README_ZH.md">中文</a> | <a href="README.md">English</a></p>
</div>

![GooseForum interface preview](https://github.com/leancodebox/assert/blob/main/gooseforum-readme-poster.webp?raw=true)

GooseForum is an open-source forum for discussions, shared knowledge, and ongoing connections. From the first topic to managing content, members, and your community’s look, it gives you the tools to build and run a community on your own server.

[Try the demo](https://gooseforum.online/) · [Download the latest release](https://github.com/leancodebox/GooseForum/releases)

## Features

- **Make room for discussion**: Markdown topics, replies, categories, and drafts keep sharing and conversations organized.
- **Stay connected**: Notifications, private messages, and user profiles help conversations continue beyond a topic.
- **Run your community**: A built-in admin console with role and permission management.
- **Make it your own**: Customize your logo, brand copy, and footer, then preview and publish light and dark themes in the theme workbench.
- **Join from any screen**: Desktop and mobile layouts, smooth navigation, and accessible pages when JavaScript is unavailable.
- **Host it yourself**: A single executable, SQLite by default, optional MySQL, and scheduled backups.

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

- Go 1.27+
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

See [configuration documentation](docs/configuration.md) for MySQL, mail, backup, security, and site settings.

### Admin Commands

```bash
./GooseForum migrate
./GooseForum set-user-admin <userId>
./GooseForum set-user-email <userId> <email>
./GooseForum set-user-password <userId> <password>
```

## Development

The backend uses Go, Gin, and GORM; the frontend uses Vue 3, TypeScript, Vite, and TailwindCSS. Server payloads drive navigation, with GoHTML fallback pages for SEO and access without JavaScript.

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
├── docs/                   # Configuration documentation
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

- [Configuration](docs/configuration.md)

## License

MIT License. See [LICENSE](LICENSE).
