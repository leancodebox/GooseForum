package i18n

import (
	"regexp"
	"slices"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":               Fallback,
		"zh":             "zh",
		"ZH":             "zh",
		"zh-CN":          "zh",
		"en":             "en",
		"en-US,en;q=0.9": "en",
		"ru-RU":          "ru",
		"ja_JP":          "ja",
		"ko-KR":          "ko",
		"  en  ":         "en",
		"fr-FR,fr;q=0.7": "fr",
	}
	for input, want := range cases {
		if got := Normalize(input); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTranslateFallback(t *testing.T) {
	// A key present in every locale returns the localized value.
	if got := T("en", "search"); got != "Search" {
		t.Errorf(`T("en","search") = %q, want "Search"`, got)
	}
	// An unsupported locale falls back to zh (the source locale).
	if got := T("nl", "search"); got != "搜索" {
		t.Errorf(`T("nl","search") = %q, want the zh fallback`, got)
	}
	// An unknown key returns the key itself (visible, never blank).
	if got := T("en", "does.not.exist"); got != "does.not.exist" {
		t.Errorf(`T for missing key = %q, want the key`, got)
	}
}

func TestNewLocalesRenderLocalizedSearch(t *testing.T) {
	for lang, want := range map[string]string{"ko": "검색", "fr": "Rechercher"} {
		if got := T(lang, "search"); got != want {
			t.Errorf("T(%q, search) = %q, want %q", lang, got, want)
		}
	}
}

func TestTranslateInterpolation(t *testing.T) {
	got := T("en", "meta.searchDesc", "site", "GooseForum")
	want := "Search topics, keywords, and discussions on GooseForum."
	if got != want {
		t.Errorf("interpolated = %q, want %q", got, want)
	}
	// Numeric params render without quotes.
	if got := T("en", "partnersSummary", "count", 7); got != "7 community partners." {
		t.Errorf("partnersSummary = %q", got)
	}
}

func TestFunc(t *testing.T) {
	tr := Func("ja")
	if got := tr("search"); got != "検索" {
		t.Errorf(`Func("ja")("search") = %q, want "検索"`, got)
	}
}

func TestNewLocalesHaveCompleteDictionaries(t *testing.T) {
	placeholder := regexp.MustCompile(`\{\w+\}`)
	for _, lang := range []string{"ko", "fr"} {
		if len(dictionaries[lang]) != len(dictionaries[Fallback]) {
			t.Errorf("%s dictionary has %d entries, want %d", lang, len(dictionaries[lang]), len(dictionaries[Fallback]))
		}
		for key, original := range dictionaries[Fallback] {
			translated := dictionaries[lang][key]
			if translated == "" {
				t.Errorf("%s: missing text for %s", lang, key)
				continue
			}
			want := placeholder.FindAllString(original, -1)
			got := placeholder.FindAllString(translated, -1)
			slices.Sort(want)
			slices.Sort(got)
			if !slices.Equal(got, want) {
				t.Errorf("%s %s: placeholders = %v, want %v", lang, key, got, want)
			}
		}
	}
}
