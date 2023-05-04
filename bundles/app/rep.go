package app

import "embed"

var actorFS embed.FS
var defaultConfig string

func ActorSave(dataWebFS embed.FS) {
	actorFS = dataWebFS
}

func GetActorFS() embed.FS {
	return actorFS
}

func ConfigData(data string) {
	defaultConfig = data
}

func GetDefaulConfig() string {
	return defaultConfig
}
