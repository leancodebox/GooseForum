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

//go:embed default-avatar.png
var defaultAvatar []byte

func GetDefaultAvatar() []byte {
	return defaultAvatar
}
