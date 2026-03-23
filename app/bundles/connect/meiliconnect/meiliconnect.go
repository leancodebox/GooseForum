package meiliconnect

import (
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/meilisearch/meilisearch-go"
)

var (
	client = getClient()
)

func getClient() meilisearch.ServiceManager {
	url := preferences.Get("meilisearch.url")
	if url == "" {
		return nil
	}
	if preferences.Get("meilisearch.masterkey") != "" {
		key := meilisearch.WithAPIKey(preferences.Get("meilisearch.masterkey"))
		return meilisearch.New(url, key)
	}

	return meilisearch.New(url)
}

func GetClient() meilisearch.ServiceManager {
	return client
}

// IsAvailable 检查 Meilisearch 是否可用
func IsAvailable() bool {
	if client == nil {
		return false
	}
	_, err := client.Health()
	return err == nil
}
