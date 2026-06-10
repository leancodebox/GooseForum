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
				Tokens: pageConfig.SiteThemeTokens{
					ColorBase100: "oklch(100% 0 0)",
					RadiusField:  "6px",
				},
			},
		},
		Draft: &pageConfig.SiteThemeSnapshot{
			Enabled: true,
			Themes: []pageConfig.SiteThemeDefinition{
				{
					Name:        "gf-dark",
					ColorScheme: "dark",
					Tokens: pageConfig.SiteThemeTokens{
						ColorBase100: "oklch(20% 0 0)",
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
						Tokens: pageConfig.SiteThemeTokens{
							ColorBase100: "oklch(99% 0 0)",
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
	if normalized.Themes[0].Tokens.RadiusField != "0.5rem" {
		t.Fatalf("expected radius-field legacy value to be normalized, got %q", normalized.Themes[0].Tokens.RadiusField)
	}
}
