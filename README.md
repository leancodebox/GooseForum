<h1  align="center">
  <br>
  <a href="https://github.com/leancodebox/GooseForum" alt="logo" >
    <img src="resource/static/pic/default-avatar.png" width="140"/></a>
  <br>
  GooseForum
  <br>
</h1>


GooseForum 是一个现代化的论坛系统，采用 Vue 3 + Go 的前后端分离架构。[站点体验](https://gooseforum.online/)

## 快速安装

[releases](https://github.com/leancodebox/GooseForum/releases) 下载合适系统版本的压缩包。解压后通过运行 `./GooseForum serve` 即可启动 `GooseForum` 管理员账号默认 username `admin` password `gooseforum`  
默认管理地址 http://localhost:99/app/admin  

## 技术栈

### 前端
- Vue 3 (Composition API)
- Vue Router
- Naive UI
- Vite
- JavaScript/ES6+

### 后端
- Go
- Gin Framework

## 项目结构

## 编译相关

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
