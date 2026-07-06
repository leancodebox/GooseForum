package i18n

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":               Fallback,
		"zh":             "zh",
		"ZH":             "zh",
		"zh-CN":          "zh",
		"en":             "en",
		"en-US,en;q=0.9": "en",
		"ja_JP":          "ja",
		"ko":             Fallback, // unsupported -> fallback
		"  en  ":         "en",
		"fr-FR,fr;q=0.7": Fallback,
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
	if got := T("ko", "search"); got != "搜索" {
		t.Errorf(`T("ko","search") = %q, want the zh fallback`, got)
	}
	// An unknown key returns the key itself (visible, never blank).
	if got := T("en", "does.not.exist"); got != "does.not.exist" {
		t.Errorf(`T for missing key = %q, want the key`, got)
	}
}

func TestTranslateInterpolation(t *testing.T) {
	got := T("en", "searchSummary", "query", "goose", "total", 42)
	want := "42 results for “goose”."
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
