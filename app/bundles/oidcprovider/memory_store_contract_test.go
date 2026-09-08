package oidcprovider_test

import (
	"testing"

	provider "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/bundles/oidcprovider/storetest"
)

func TestMemoryStoreContract(t *testing.T) {
	storetest.Run(t, func(*testing.T) (provider.Store, func(*provider.Client)) {
		store := provider.NewMemoryStore()
		return store, store.PutClient
	})
}
