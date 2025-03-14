# GooseForum

GooseForum 是一个现代化的论坛系统，采用 Vue 3 + Go 的前后端分离架构。

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
