package assert

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"text/template"
)

//go:embed config.templ.toml
var configTempl []byte

func GetDefaultConfig() []byte {
	return configTempl
}

func GenerateConfig() []byte {
	var b bytes.Buffer
	t := template.New("config.templ.toml")
	t = template.Must(t.Parse(string(configTempl)))
	err := t.Execute(&b, map[string]any{
		"SigningKey": algorithm.SafeGenerateSigningKey(32),
	})

	if err != nil {
		fmt.Println(err)
	}
	return b.Bytes()
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
