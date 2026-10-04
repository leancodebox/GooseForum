package api

import (
	"fmt"
	"os"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/service/kvstore"
)

func TestMain(m *testing.M) {
	path, err := os.MkdirTemp("", "gooseforum-api-test-kv-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// KV initializes once per process, so isolate it before any test can open it.
	preferences.Set("badger.path", path)
	code := m.Run()
	kvstore.Close()
	if err := os.RemoveAll(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}
