package forum

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestSiteHeaderPayloadKeepsDisabledLinksHidden(t *testing.T) {
	if items := buildHeaderNavItems(nil); len(items) != 2 {
		t.Fatalf("legacy configuration has %d default items, want 2", len(items))
	}

	header := defaultconfig.GetDefaultSiteChromeHeader()
	header[0].Enabled = false
	items := buildHeaderNavItems(header)
	if len(items) != 1 || items[0].Key != "links" {
		t.Fatalf("visible items = %+v, want links only", items)
	}

	header[1].Enabled = false
	raw, err := json.Marshal(LayoutPayload{Header: buildHeaderNavItems(header)})
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Header json.RawMessage `json:"header"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if string(payload.Header) != "[]" {
		t.Fatalf("disabled header must be an explicit empty array, got %q", payload.Header)
	}
}

func TestSidebarVisibilityIsIndependentOfHeader(t *testing.T) {
	var chrome pageConfig.SiteChromeConfig
	if err := json.Unmarshal([]byte(`{"header":[{"id":"sponsors","enabled":false}]}`), &chrome); err != nil {
		t.Fatal(err)
	}
	if sidebar := buildSidebarPayload(chrome, nil, "topics"); len(sidebar.HiddenSystemLinks) != 0 {
		t.Fatalf("legacy header configuration hides sidebar links: %v", sidebar.HiddenSystemLinks)
	}

	if err := json.Unmarshal([]byte(`{"hiddenSidebarSystemLinks":["members","links"]}`), &chrome); err != nil {
		t.Fatal(err)
	}
	sidebar := buildSidebarPayload(chrome, nil, "topics")
	if !slices.Equal(sidebar.HiddenSystemLinks, []string{"members", "links"}) {
		t.Fatalf("hidden system links = %v, want [members links]", sidebar.HiddenSystemLinks)
	}
}
