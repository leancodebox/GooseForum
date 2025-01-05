package assert

import (
	"embed"
	_ "embed"
)

//go:embed config.example.toml
var configExample []byte

func GetDefaultConfig() []byte {
	return configExample
}

//go:embed  all:frontend/**
var actorFS embed.FS

func GetActorFs() embed.FS {
	return actorFS
}

//go:embed 07akioni.jpeg
var defaultAvatar []byte

func GetDefaultAvatar() []byte {
	return defaultAvatar
}
