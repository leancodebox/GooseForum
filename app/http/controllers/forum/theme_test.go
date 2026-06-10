package forum

import (
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestNormalizeSiteThemeConfigDoesNotMutateInput(t *testing.T) {
	input := pageConfig.SiteThemeConfig{
		Enabled: true,
		Themes: []pageConfig.SiteThemeDefinition{
			{
				Name:        "gf-light",
				ColorScheme: "light",
				Tokens: map[string]string{
					"color-base-100": "oklch(100% 0 0)",
					"unknown-token":  "remove-me",
				},
			},
		},
		Draft: &pageConfig.SiteThemeSnapshot{
			Enabled: true,
			Themes: []pageConfig.SiteThemeDefinition{
				{
					Name:        "gf-dark",
					ColorScheme: "dark",
					Tokens: map[string]string{
						"color-base-100": "oklch(20% 0 0)",
					},
				},
			},
		},
		History: []pageConfig.SiteThemeSnapshot{
			{
				Enabled: true,
				Themes: []pageConfig.SiteThemeDefinition{
					{
						Name:        "gf-light",
						ColorScheme: "light",
						Tokens: map[string]string{
							"color-base-100": "oklch(99% 0 0)",
						},
					},
				},
			},
		},
	}
	original := cloneSiteThemeConfig(input)

	normalized := normalizeSiteThemeConfig(input)

	if !reflect.DeepEqual(input, original) {
		t.Fatal("normalizeSiteThemeConfig should not mutate the input config")
	}
	if _, ok := normalized.Themes[0].Tokens["unknown-token"]; ok {
		t.Fatal("normalized theme should drop unknown tokens")
	}
}
