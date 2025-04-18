<h1  align="center">
  <br>
  <a href="https://github.com/leancodebox/GooseForum" alt="logo" >
    <img src="resource/static/pic/default-avatar.png" width="140"/></a>
  <br>
  GooseForum
  <br>
</h1>


GooseForum 是一个现代化的论坛系统，采用 Vue 3 + Go 的前后端分离架构。提供了极低依赖的安装方式。

如果你向知道GooseForum运行效果怎么样，请点击[GooseForum](https://gooseforum.online/)进行体验。

# GooseForum 快速上手

[在GooseForum上查看快手上手](https://gooseforum.online/post/371)

## 获取 GooseForum

你可以在 [GitHub Release](https://github.com/leancodebox/GooseForum/releases) 页面获取已经构建打包完成的主程序。其中每个版本都提供了常见系统架构下可用的主程序，命名规则为`GooseForum_操作系统_CPU架构.tar.gz` 。比如，普通 64 位 Linux 系统上部署 v0.0.2 版本，则应该下载GooseForum_Linux_x86_64.tar.gz。

## 启动 GooseForum

```shell
#解压获取到的主程序
tar -zxvf GooseForum_OS_ARCH.tar.gz

# 赋予执行权限
chmod +x ./GooseForum

# 启动 GooseForum
./GooseForum serve
```

GooseForum 在首次启动时，会创建初始管理员账号，管理员账号为`admin`、密码为`gooseforum`。如果你在个人中心更改了密码后忘记管理员密码，你可以通过执行`./GooseForum user:changePassword --userId=1 --password=q1234567890`命令进行重置。

GooseForum 默认会监听99端口。你可以在浏览器中访问 http://服务器IP:99 进入 GooseForum。

如果你需要进入管理员页面可以通过访问  http://服务器IP:99/app/admin 进入

以上步骤操作完后，最简单的部署就完成了。你可能需要一些更为具体的配置，才能让 GooseForum 更好的工作，具体流程请参考下面的配置流程。

## 配置文件 

GooseForum 启动总会默认检查执行的同级目录是否存在`config.toml`，如果不存在则会进行创建，同时使用本文件进行项目启动。默认情况下，你不需要更改任何配置。如果有需要你可以参考下方相关配置文件解释

```toml
[app]
name = "app"
env = "production" # APP_ENV in local,production 会影响某些加载逻辑，生产环境不要更改
debug = false # 日志会更详细一般不用调整

[server]
url = "http://localhost" # 影响一些地址返回的url ，例如 rss sitemap
port = 99 # 启动端口

[footer]
url = "https://github.com/leanCodeBox/GooseForum" # 项目
text = "Powered by GooseForum"

[jwtopt]
signingKey="signingKey" # 项目生成的为一个随机 signingKey 你和别人的不会一样，一般情况不用修改，修改的话，会导致已经登录的用户退出登录。
validTime = 604800

[mail]
host = "smtp.example.com" # 邮箱相关配置，可以用来邮件激活
port = 587
username = "noreply@example.com"
password = "your-password"
from_name = "GooseForum"

[db]
migration = "on" # on,off # 数据库迁移 ，如果你没有进行GooseForum 版本替换，可以启动一次后调整为 off
backupSqlite = true # 是否定时备份sqlite数据库 
backupDir = "./storage/databasebackup/" # 备份地址
keep = 7 # 备份数量
spec = "0 3 * * *" # cron定时 ，这一句是每天凌晨3点更新，默认不用调整

[db.default]
connection = "sqlite"# in mysql sqlite # 默认使用sqlite 项目使用wal，一般不用调整也可以
url = "db_user:db_pass@tcp(db_host:3306)/db_name?charset=utf8mb4&parseTime=True&loc=Local"
path = "./storage/database/sqlite.db"# :memory:|./storage/database/sqlite.db

maxIdleConnections = 3
maxOpenConnections = 5
maxLifeSeconds = 300

[db.file]
connection = "sqlite"# in mysql sqlite
url = "root:root_password@tcp(127.0.0.1:3306)/goose_forum?charset=utf8mb4&parseTime=True&loc=Local"
path = "./storage/database/file.db"# :memory:|./storage/database/sqlite.db

maxIdleConnections = 3
maxOpenConnections = 5
maxLifeSeconds = 300


[log]
type = "file"# LOG_TYPE stdout,file 日志输出格式，如果使用 stdout 会在控制台输出
path = "./storage/logs/run.log"
rolling = false # 是否开启滚动 true ,false
maxage = 10 # 最大日期
maxsize = 256 # 最大文件大小 MB
maxBackUps = 30 # 最大保留文件数量

[site] # html 中 meta 扩展内容，如果你有baidu等验证网站所有权的需要，可以更改此处
metaList="""
[{"name":"author","content":"GooseForum's Friend"}] 
"""
```

## 构建

### 环境准备

- 参照 [Getting Started - The Go Programming Language](https://go.dev/doc/install)  安装并配置 Go 语言开发环境 (>=1.18)；
- 参考 [下载|Node.js](https://nodejs.org/zh-cn/download/) 安装 Node.js;


### 开始构建

#### 克隆代码

```shell
git clone git@github.com:leancodebox/GooseForum.git
cd GooseForum
```

#### 构建项目前后端分离资源并编译完整项目为二进制可执行文件

GooseForum 项目主要由两部分组成：二者均在同一仓库，分别为主目录下的服务端和`actorv2`目录下的前后端分离项目，需要先构建`actorv2` 目录下的前后端分离项目。完整命令如下

```shell
cd actorv2
npm i
npm run build
cd .. 
go mod tidy
go build 
```

编译完成后，会在项目根目录下生成最终的可执行文件 `GooseForum` 。

#### 构建助手

你可以使用 goreleaser 快速完成构建、打包等操作，使用方法如下：

##### 安装 goreleaser

```shell
go install github.com/goreleaser/goreleaser@latest
```

##### 构建项目

```shell
goreleaser build --clean --single-target --snapshot
```

或者交叉编译出所有可用版本：
```shell
goreleaser build --clean --snapshot
```

------------------------------------------

### 技术栈

#### 前端
- Vue 3 (Composition API)
- Vue Router
- Naive UI
- Vite
- JavaScript/ES6+

#### 后端
- Go
- Gin Framework

### 项目结构

### 编译相关

```
 go generate ./...
```
```
go build -ldflags="-w -s" .
```
windows
```
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build

SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=arm64
go build

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build
```

powershell
```powershell
#设置Linux编译环境
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
 
$env:CGO_ENABLED=""
$env:GOOS=""
$env:GOARCH=""
# 开始编译
go build .

#https://learn.microsoft.com/zh-cn/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-5.1
```

mac
```
go build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
```

linux
```
go build
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
```

# Todo

[TodoList](./TodoList.md)
