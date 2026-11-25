package forum

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
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

func TestSidebarCategoryPayloadCarriesSavedIcons(t *testing.T) {
	sidebar := buildSidebarPayload(pageConfig.SiteChromeConfig{}, []*category.Entity{
		{Id: 1, Name: "Coding", Slug: "coding", Color: "#41d641", Icon: "lucide:book-open"},
		{Id: 2, Name: "General", Slug: "general", Color: "#8241d6"},
	}, "topics")
	if len(sidebar.Categories) != 2 || sidebar.Categories[0].Icon != "lucide:book-open" ||
		sidebar.Categories[1].Icon != "" || sidebar.Categories[1].Color != "#8241d6" {
		t.Fatalf("unexpected sidebar categories: %+v", sidebar.Categories)
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

func TestPageIconsIncludesOnlyReferencedLucideDefinitions(t *testing.T) {
	page := PagePayload{
		Layout: LayoutPayload{
			Header: []NavItemPayload{{Icon: "book-open"}},
			Sidebar: SidebarPayload{
				Groups:     []SidebarGroupPayload{{Items: []NavItemPayload{{Icon: "book-open"}, {Icon: "unknown-icon"}}}},
				Resources:  []NavItemPayload{{Icon: "alarm-check"}},
				Categories: []CategoryNavPayload{{Icon: "lucide:code"}, {Icon: "🎉"}, {Icon: "/avatar.png"}},
			},
		},
		Props: CategoriesPageProps{Categories: []CategoryDirectoryPayload{{Icon: "lucide:star"}}},
	}
	result := pageIcons(page)
	if len(result) != 4 {
		t.Fatalf("got %d definitions, want 4", len(result))
	}
	for _, name := range []string{"book-open", "code", "alarm-check", "star"} {
		var icon struct{ Node []json.RawMessage }
		if err := json.Unmarshal(result[name], &icon); err != nil || len(icon.Node) == 0 {
			t.Fatalf("missing SVG definition for %q", name)
		}
	}
}
