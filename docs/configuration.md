# GooseForum 配置文档

本文档详细说明了 GooseForum 的所有配置选项。GooseForum 启动时会自动检查执行目录下是否存在 `config.toml` 文件，如果不存在则会自动创建一个默认配置文件。

## 📋 配置文件结构

配置文件采用 TOML 格式，包含以下主要部分：

- [app] - 应用基础配置
- [server] - 服务器配置
- [jwtopt] - JWT 认证配置
- [db] - 数据库配置
- [log] - 日志配置
- [site] - 站点元数据配置

## 🔧 详细配置说明

### [app] 应用配置

```toml
[app]
name = "app"                    # 应用名称
env = "production"              # 运行环境: local, production
debug = false                   # 调试模式，开启后日志更详细
maintenance = false             # 维护模式，开启后显示维护页面
```

**配置说明：**
- `env`: 影响某些加载逻辑，生产环境建议保持 `production`
- `debug`: 开启后会输出更详细的日志信息，用于开发调试
- `maintenance`: 开启后网站将显示维护页面，用于系统维护

### [server] 服务器配置

```toml
[server]
url = "http://localhost"        # 站点基础URL
port = 99                       # 监听端口
```

**配置说明：**
- `url`: 影响 RSS、Sitemap 等功能返回的 URL 地址
- `port`: 服务监听端口，默认 99

### [jwtopt] JWT 认证配置

```toml
[jwtopt]
signingKey = "your-random-signing-key"  # JWT 签名密钥
validTime = 604800                      # Token 有效期（秒）
```

**配置说明：**
- `signingKey`: JWT 签名密钥，系统会自动生成随机密钥。**注意：修改此值会导致所有已登录用户退出登录**
- `validTime`: Token 有效期，默认 604800 秒（7天）


### [db] 数据库配置

```toml
[db]
migration = "on"                           # 数据库迁移: on, off
backupSqlite = true                        # 是否定时备份 SQLite 数据库
backupDir = "./storage/databasebackup/"    # 备份目录
keep = 7                                   # 保留备份数量
spec = "0 3 * * *"                         # 备份时间（Cron 表达式）
```

**配置说明：**
- `migration`: 数据库迁移开关，首次启动或版本升级时建议开启
- `backupSqlite`: 是否启用 SQLite 自动备份
- `backupDir`: 备份文件存储目录
- `keep`: 保留的备份文件数量
- `spec`: Cron 表达式，默认每天凌晨 3 点备份

#### [db.default] 主数据库配置

```toml
[db.default]
connection = "sqlite"                                                           # 数据库类型: sqlite, mysql
url = "db_user:db_pass@tcp(db_host:3306)/db_name?charset=utf8mb4&parseTime=True&loc=Local"  # MySQL 连接字符串
path = "./storage/database/sqlite.db"                                          # SQLite 数据库文件路径

maxIdleConnections = 3                                                          # 最大空闲连接数
maxOpenConnections = 5                                                          # 最大打开连接数
maxLifeSeconds = 300                                                            # 连接最大生存时间（秒）
```

**SQLite 配置示例：**
```toml
[db.default]
connection = "sqlite"
path = "./storage/database/sqlite.db"
maxIdleConnections = 3
maxOpenConnections = 5
maxLifeSeconds = 300
```

**MySQL 配置示例：**
```toml
[db.default]
connection = "mysql"
url = "username:password@tcp(localhost:3306)/gooseforum?charset=utf8mb4&parseTime=True&loc=Local"
maxIdleConnections = 10
maxOpenConnections = 20
maxLifeSeconds = 3600
```

#### [db.file] 文件数据库配置

```toml
[db.file]
connection = "sqlite"                                                           # 数据库类型
url = "root:root_password@tcp(127.0.0.1:3306)/goose_forum?charset=utf8mb4&parseTime=True&loc=Local"  # MySQL 连接字符串
path = "./storage/database/file.db"                                            # SQLite 数据库文件路径

maxIdleConnections = 3
maxOpenConnections = 5
maxLifeSeconds = 300
```

### [log] 日志配置

```toml
[log]
type = "file"                   # 日志输出类型: stdout, file
path = "./storage/logs/run.log" # 日志文件路径
rolling = true                  # 是否开启日志滚动
maxage = 10                     # 日志文件最大保存天数
maxsize = 256                   # 单个日志文件最大大小（MB）
maxBackUps = 30                 # 最大保留日志文件数量
```

**配置说明：**
- `type`: 
  - `stdout`: 输出到控制台
  - `file`: 输出到文件
- `rolling`: 开启后会自动切割日志文件
- `maxage`: 超过指定天数的日志文件会被自动删除
- `maxsize`: 单个日志文件超过指定大小后会自动切割
- `maxBackUps`: 保留的历史日志文件数量


## 🔄 配置文件热重载

GooseForum 支持配置文件热重载，修改 `config.toml` 文件后无需重启服务即可生效（部分配置除外）。

**需要重启的配置项：**
- 服务器端口 (`server.port`)
- 数据库连接配置 (`db.*`)
- 日志配置 (`log.*`)

## 🛡 安全建议

1. **JWT 签名密钥**：使用强随机密钥，不要使用默认值
2. **数据库密码**：使用复杂密码，定期更换
3. **邮箱密码**：使用应用专用密码或授权码
4. **文件权限**：确保配置文件权限设置正确（建议 600）

```bash
# 设置配置文件权限
chmod 600 config.toml
```

## 🔍 故障排除

### 常见问题

1. **服务启动失败**
   - 检查端口是否被占用
   - 检查配置文件语法是否正确
   - 查看日志文件获取详细错误信息

2. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接字符串是否正确
   - 确认数据库用户权限

3. **邮件发送失败**
   - 检查 SMTP 服务器配置
   - 验证邮箱用户名和密码
   - 确认网络连接正常

### 调试模式

开启调试模式获取更多信息：

```toml
[app]
debug = true

[log]
type = "stdout"
```

## 📚 相关文档

- [快速开始](../README.md#🚀-快速开始)