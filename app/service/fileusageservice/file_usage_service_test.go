package fileusageservice

import "testing"

func TestFileNameFromURL(t *testing.T) {
	tests := map[string]string{
		"/file/img/2026/06/a.webp":              "2026/06/a.webp",
		"https://example.com/file/img/a/b.webp": "a/b.webp",
		"avatars/1/avatar.webp":                 "avatars/1/avatar.webp",
		"/static/pic/default-avatar.webp":       "",
		"https://example.com/static/a.webp":     "",
		"../secret.webp":                        "",
	}
	for input, want := range tests {
		if got := fileNameFromURL(input); got != want {
			t.Fatalf("fileNameFromURL(%q) = %q, want %q", input, got, want)
		}
	}
}
