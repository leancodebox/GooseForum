# GooseForum Configuration

[中文](configuration_ZH.md) | [English](configuration.md)

This document describes the main GooseForum configuration options. On startup, GooseForum checks for `config.toml` in its working directory and creates a default configuration if the file does not exist.

## Configuration File Structure

The configuration uses TOML and contains these main sections:

- `[app]`: Application settings
- `[server]`: Server settings
- `[jwtopt]`: JWT authentication
- `[db]`: Database settings
- `[log]`: Logging

## Configuration Options

### [app] Application Settings

```toml
[app]
env = "production"              # Environment: local, production
debug = false                   # Optional debug override
cdn_url = ""                    # CDN URL
```

- `env`: Controls environment-dependent loading behavior. Use `production` for production deployments.
- `debug`: Optional override. When omitted, debug mode defaults to enabled in `local` and disabled in other environments. Enabling it provides more detailed log output.
- `cdn_url`: CDN URL for static assets.

### [server] Server Settings

```toml
[server]
url = "http://localhost"        # Base site URL
port = 5234                     # Listening port
```

- `url`: Used to generate URLs for features such as RSS and sitemaps.
- `port`: The listening port; defaults to 5234.

### [jwtopt] JWT Authentication

```toml
[jwtopt]
signingKey = "your-random-signing-key"  # JWT signing key
validTime = 604800                      # Token lifetime in seconds
```

- `signingKey`: GooseForum automatically generates a random signing key. **Changing this value signs out all currently logged-in users.**
- `validTime`: Token lifetime; defaults to 604800 seconds (7 days).

### [db] Database Settings

```toml
[db]
migration = "on"                         # Database migrations: on, off
backupSqlite = true                      # Enable scheduled SQLite backups
backupDir = "./storage/databasebackup/"  # Backup directory
keep = 7                                 # Number of backups to retain
spec = "0 3 * * *"                       # Backup schedule (cron expression)
```

- `migration`: Enables database migrations. Keep it enabled for initial setup and version upgrades.
- `backupSqlite`: Enables automatic SQLite backups.
- `backupDir`: Directory for backup files.
- `keep`: Number of backup files to retain.
- `spec`: Cron expression; the default runs backups daily at 03:00.

#### [db.default] Primary Database

```toml
[db.default]
connection = "sqlite"                # Database type: sqlite, mysql
path = "./storage/database/sqlite.db" # SQLite database path
url = "user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local" # MySQL DSN
maxIdleConnections = 3               # Maximum idle connections
maxOpenConnections = 5               # Maximum open connections
maxLifeSeconds = 300                 # Maximum connection lifetime in seconds
```

SQLite example:

```toml
[db.default]
connection = "sqlite"
path = "./storage/database/sqlite.db"
```

MySQL example:

```toml
[db.default]
connection = "mysql"
url = "username:password@tcp(localhost:3306)/gooseforum?charset=utf8mb4&parseTime=True&loc=Local"
maxIdleConnections = 10
maxOpenConnections = 20
maxLifeSeconds = 3600
```

### [log] Logging

```toml
[log]
type = "stdout"                 # Output: stdout, file
path = "./storage/logs/run.log" # Log file path
rolling = true                  # Enable log rotation
maxage = 10                     # Maximum retention in days
maxsize = 256                   # Maximum file size in MB
maxBackUps = 30                 # Maximum number of backup log files
```

- `type`: `stdout` writes to the console; `file` writes to a file.
- `rolling`: Enables automatic log rotation.
- `maxage`: Rotated log files older than this number of days are removed.
- `maxsize`: Log files are rotated when they exceed this size.
- `maxBackUps`: Number of rotated log files to retain.

### OAuth Login

OAuth settings are stored in the database and managed under **Settings > OAuth** in the admin console. Built-in providers include GitHub, Google, and Discord. You can also add OpenID Connect providers that support Discovery.

The admin console displays the callback URL for each provider. A provider appears on the login and account linking pages only when it is enabled and both its Client ID and Client Secret are configured. Legacy `[github]` settings are migrated into the database on startup; remove them from `config.toml` after confirming that migration succeeded.

## Configuration Reloading

GooseForum supports configuration reloading. Changes to `config.toml` take effect without a restart, except for settings that require one.

Restart the service after changing:

- The listening port (`server.port`)
- Database connection settings (`db.*`)
- Logging settings (`log.*`)

## Security Recommendations

1. **JWT signing key**: Use a strong random key rather than the example placeholder.
2. **Database password**: Use a strong password and rotate it regularly.
3. **Mail password**: Use an application-specific password or authorization code.
4. **File permissions**: Restrict access to the configuration file; mode `600` is recommended.

```bash
chmod 600 config.toml
```

## Troubleshooting

### Common Problems

**The service does not start**

- Check whether the port is already in use.
- Check the configuration file for syntax errors.
- Read the logs for detailed error messages.

**The database connection fails**

- Check whether the database service is running.
- Verify the connection string.
- Check the database user's permissions.

**Email delivery fails**

- Check the SMTP settings in the admin console.
- Verify the mail username and password.
- Check network connectivity.

### Debug Mode

To explicitly enable debug mode:

```toml
[app]
env = "production"
debug = true

[log]
type = "stdout"
```

## Related Documentation

- [Quick Start](../README.md#quick-start)
- [Chinese Configuration Guide](configuration_ZH.md)
