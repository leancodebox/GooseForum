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
	header[1].I18nLabel = ""
	items := buildHeaderNavItems(header)
	if len(items) != 1 || items[0].Key != "links" {
		t.Fatalf("visible items = %+v, want links only", items)
	}
	if items[0].I18nLabel != "shell.nav.links" {
		t.Fatalf("system link lost its translation: %+v", items[0])
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

func TestChromeNavPayloadPreservesOptionalIcon(t *testing.T) {
	var chrome pageConfig.SiteChromeConfig
	if err := json.Unmarshal([]byte(`{"resources":[
		{"id":"docs","enabled":true,"label":"Docs","url":"/docs","icon":"book-open"},
		{"id":"legacy","enabled":true,"label":"Legacy","url":"/legacy"}
	]}`), &chrome); err != nil {
		t.Fatal(err)
	}
	sidebar := buildSidebarPayload(chrome, nil, "topics")
	if len(sidebar.Resources) != 2 || sidebar.Resources[0].Icon != "book-open" || sidebar.Resources[1].Icon != "" {
		t.Fatalf("resource icons did not survive config decoding: %+v", sidebar.Resources)
	}
	raw, err := json.Marshal(sidebar.Resources[1])
	if err != nil {
		t.Fatal(err)
	}
	var legacy map[string]json.RawMessage
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatal(err)
	}
	if _, exists := legacy["icon"]; exists {
		t.Fatalf("legacy resource unexpectedly includes an icon: %s", raw)
	}
}
