package main

import (
	_ "github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/console"
	_ "net/http/pprof"
)

//go:generate npm run --prefix actor build --emptyOutDir
func main() {
	// 注册静态资源
	console.Execute()
}
