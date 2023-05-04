init

git fetch --all &&  git reset --hard origin/master && git pull

```
Mac 下编译 Linux 和 Windows 64位可执行程序

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build

linux 下编译 Mac 和 Windows 64位可执行程序

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build

Windows 下编译 Mac 和 Linux 64位可执行程序

SET CGO_ENABLED=0SET GOOS=darwinSET GOARCH=amd64 go build main.go
SET CGO_ENABLED=0SET GOOS=linuxSET GOARCH=amd64 go build main.go
```





