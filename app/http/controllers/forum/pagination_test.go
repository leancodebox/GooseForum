package forum

import "testing"

func TestBuildPagePaginationOmitsNextURLWhenExhausted(t *testing.T) {
	called := false
	got := buildPagePagination(3, false, func(page int) string {
		called = true
		return "?page=" + string(rune('0'+page))
	})
	if got.Page != 3 || got.NextPage != 0 || got.HasNext || got.NextURL != "" {
		t.Fatalf("pagination = %+v, want exhausted page with empty next URL", got)
	}
	if called {
		t.Fatal("next URL builder called for exhausted page")
	}
}

func TestBuildPagePaginationBuildsNextURLOnlyWhenAvailable(t *testing.T) {
	got := buildPagePagination(3, true, func(page int) string {
		return "/topics?page=" + string(rune('0'+page))
	})
	if got.Page != 3 || got.NextPage != 4 || !got.HasNext || got.NextURL != "/topics?page=4" {
		t.Fatalf("pagination = %+v, want page 4 URL", got)
	}
}
