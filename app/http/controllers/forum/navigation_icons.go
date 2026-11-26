package forum

import (
	"encoding/json"
	"strings"

	"github.com/leancodebox/GooseForum/resource/icons"
)

func pageIcons(page PagePayload) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage)
	switch page.Component {
	case PageComponentLogin, PageComponentResetPassword, PageComponentOIDCConsent:
		return result
	}
	add := func(name string) {
		if icon := icons.Lookup(name); icon != nil {
			result[name] = icon
		}
	}
	addMenu := func(items []NavItemPayload) {
		for _, item := range items {
			add(item.Icon)
		}
	}
	addCategory := func(value string) {
		if name, ok := strings.CutPrefix(value, "lucide:"); ok {
			add(name)
		}
	}
	sidebar := page.Layout.Sidebar
	addMenu(page.Layout.Header)
	if page.Component == PageComponentAdmin {
		return result
	}
	addMenu(sidebar.Main)
	addMenu(sidebar.Resources)
	for _, group := range sidebar.Groups {
		addMenu(group.Items)
	}
	for _, category := range sidebar.Categories {
		addCategory(category.Icon)
	}
	switch props := page.Props.(type) {
	case CategoryPageProps:
		addCategory(props.Category.Icon)
	case CategoriesPageProps:
		for _, category := range props.Categories {
			addCategory(category.Icon)
		}
	}
	return result
}
