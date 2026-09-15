package pageConfig

import (
	"fmt"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

func TestCompareAndSwapConfigRejectsStaleValue(t *testing.T) {
	if err := dbconnect.Connect().AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate page config: %v", err)
	}
	pageType := fmt.Sprintf("test-cas-%d", time.Now().UnixNano())
	if err := SaveConfig(pageType, `{"version":1}`); err != nil {
		t.Fatalf("save initial config: %v", err)
	}
	t.Cleanup(func() { builder().Where("page_type = ?", pageType).Delete(&Entity{}) })

	initial := GetByPageType(pageType)
	saved, err := CompareAndSwapConfig(initial, `{"version":2}`)
	if err != nil || !saved {
		t.Fatalf("first compare-and-swap = %v, %v", saved, err)
	}
	saved, err = CompareAndSwapConfig(initial, `{"version":3}`)
	if err != nil {
		t.Fatalf("stale compare-and-swap: %v", err)
	}
	if saved {
		t.Fatal("stale compare-and-swap overwrote current config")
	}
	if got := GetByPageType(pageType).Config; got != `{"version":2}` {
		t.Fatalf("config = %s", got)
	}
}
