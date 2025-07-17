package meiliconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/meilisearch/meilisearch-go"
)

var client = getClient()

func getClient() meilisearch.ServiceManager {
	url := preferences.Get("meilisearch.url")
	if preferences.Get("meilisearch.masterkey") != "" {
		key := meilisearch.WithAPIKey(preferences.Get("meilisearch.masterkey"))
		return meilisearch.New(url, key)
	} else {
		return meilisearch.New(url)
	}
}

func GetClient() meilisearch.ServiceManager {
	return client
}
