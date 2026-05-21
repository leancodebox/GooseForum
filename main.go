package main

import (
	_ "net/http/pprof"

	_ "github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/console"
)

// --go:generate go run generatetool/generatetool.go
//
// -- go:generate npm run --prefix actor build --emptyOutDir
//
//go:generate pnpm --dir resource build
//go:generate pnpm --dir admin build
func main() {
	// 注册静态资源
	console.Execute()
}
