package meiliconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/meilisearch/meilisearch-go"
)

var client = getClient()

func getClient() meilisearch.ServiceManager {
	url := preferences.Get("meilisearch.url")
	key := meilisearch.WithAPIKey(preferences.Get("meilisearch.masterkey"))
	return meilisearch.New(url, key)
}

func GetClient() meilisearch.ServiceManager {
	return client
}
