package themeservice

import (
	"reflect"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestNormalizeConfigDoesNotMutateInput(t *testing.T) {
	input := pageConfig.SiteThemeConfig{
		Enabled: true,
		Themes: []pageConfig.SiteThemeDefinition{
			{
				Name:        LightName,
				ColorScheme: "light",
				Tokens: pageConfig.SiteThemeTokens{
					ColorBase100: "oklch(100% 0 0)",
					RadiusField:  "6px",
				},
			},
		},
		Prepublish: &pageConfig.SiteThemePrepublish{
			Enabled: true,
			Themes: []pageConfig.SiteThemeDefinition{
				{
					Name:        DarkName,
					ColorScheme: "dark",
					Tokens: pageConfig.SiteThemeTokens{
						ColorBase100: "oklch(20% 0 0)",
					},
				},
			},
		},
	}
	original := CloneConfig(input)

	normalized := NormalizeConfig(input)

	if !reflect.DeepEqual(input, original) {
		t.Fatal("NormalizeConfig should not mutate the input config")
	}
	if normalized.Themes[0].Tokens.RadiusField != "0.5rem" {
		t.Fatalf("expected radius-field legacy value to be normalized, got %q", normalized.Themes[0].Tokens.RadiusField)
	}
}

func TestNormalizeConfigIgnoresInvalidPrepublish(t *testing.T) {
	normalized := NormalizeConfig(pageConfig.SiteThemeConfig{
		Enabled: true,
		Prepublish: &pageConfig.SiteThemePrepublish{
			Enabled: true,
		},
	})

	if normalized.Prepublish != nil {
		t.Fatal("expected empty prepublish themes to be treated as missing")
	}
}
