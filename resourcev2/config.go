package resourcev2

import "embed"

//go:embed  all:templates/**
var templates embed.FS

func GetTemplates() embed.FS {
	return templates
}
