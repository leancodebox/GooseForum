package main

import (
	"embed"
	"github.com/leancodebox/GooseForum/app/console"
	"github.com/leancodebox/GooseForum/bundles/app"
	_ "net/http/pprof"
)

//go:embed  all:actor/dist/**
var actorFS embed.FS

//go:embed config.example.toml
var configData string

func main() {
	// 注册静态资源
	app.InitStart()
	app.ActorSave(actorFS)
	app.ConfigData(configData)
	console.Execute()
}
