// Package icons looks up SVG definitions from the Lucide version used by the UI.
package icons

import (
	_ "embed"
	"encoding/json"
)

//go:generate node ../apps/web/generate-navigation-icons.mjs

//go:embed lucide.json
var catalogJSON []byte

var catalog = func() map[string]json.RawMessage {
	var data struct {
		Icons   map[string]json.RawMessage
		Aliases map[string]string
	}
	if err := json.Unmarshal(catalogJSON, &data); err != nil {
		panic(err)
	}
	for alias, name := range data.Aliases {
		data.Icons[alias] = data.Icons[name]
	}
	return data.Icons
}()

func Lookup(name string) json.RawMessage { return catalog[name] }
