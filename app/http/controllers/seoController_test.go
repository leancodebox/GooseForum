package controllers

import "testing"

func TestFaviconLocation(t *testing.T) {
	tests := []struct {
		name     string
		siteLogo string
		want     string
	}{
		{name: "configured root path", siteLogo: "/file/img/site.webp", want: "/file/img/site.webp"},
		{name: "configured absolute URL", siteLogo: "https://cdn.example.com/favicon.ico", want: "https://cdn.example.com/favicon.ico"},
		{name: "trim whitespace", siteLogo: "  /static/pic/custom.png  ", want: "/static/pic/custom.png"},
		{name: "empty uses bundled icon", siteLogo: "", want: defaultFaviconPath},
		{name: "root favicon avoids redirect loop", siteLogo: "/favicon.ico", want: defaultFaviconPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := faviconLocation(tt.siteLogo); got != tt.want {
				t.Fatalf("faviconLocation(%q) = %q, want %q", tt.siteLogo, got, tt.want)
			}
		})
	}
}
