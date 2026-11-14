package markdown2html

import (
	"strings"
	"testing"
)

func TestGalleryRenderingAndImageAnalysis(t *testing.T) {
	for _, mode := range []string{"grid", "carousel"} {
		source := "[gallery mode=\"" + mode + "\"]\n![First](/uploads/one.jpg)\n\n![Second](/uploads/two.jpg)\n[/gallery]"
		result := AnalyzeTopicContentVersion(source, 200, 1, true)
		if !strings.Contains(result.RenderedHTML, `data-gallery="`+mode+`"`) || strings.Count(result.RenderedHTML, "<img ") != 2 {
			t.Fatalf("gallery rendering: %s", result.RenderedHTML)
		}
		if len(result.ImageURLs) != 2 {
			t.Fatalf("images lost to analysis: %+v", result)
		}
		if strings.Contains(PostMarkdownToHTMLVersion(source, 0), "data-gallery") {
			t.Fatal("legacy content was reinterpreted")
		}
	}
	caption := "[gallery mode=\"grid\"]\n![One](/one.jpg) [mention user=\"12\"]@alice[/mention]\n[/gallery]"
	if html := PostMarkdownToHTMLVersion(caption, 1); !strings.Contains(html, `data-mention-user="12"`) {
		t.Fatalf("gallery inline markup: %s", html)
	}
	for _, source := range []string{
		"[gallery mode=\"invalid\"]\n![x](/x.png)\n[/gallery]",
		"```\n[gallery mode=\"grid\"]\n![x](/x.png)\n[/gallery]\n```",
		"[future]\n[gallery mode=\"grid\"]\n![x](/x.png)\n[/gallery]\n[/future]",
	} {
		if html := PostMarkdownToHTMLVersion(source, 1); strings.Contains(html, "data-gallery") {
			t.Fatalf("unexpected gallery: %s", html)
		}
	}
}
