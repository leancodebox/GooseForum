package markdown2html

import (
	"strings"
	"testing"
)

func TestCommentMarkdownToHTMLNormalizesLinksAndImages(t *testing.T) {
	html := CommentMarkdownToHTML(`[Goose](https://gooseforum.online)

![logo](https://gooseforum.online/logo.png)`)

	if !strings.Contains(html, `target="_blank"`) {
		t.Fatalf("expected links to open in a new tab, got %s", html)
	}
	if !strings.Contains(html, `rel="nofollow ugc noopener noreferrer"`) {
		t.Fatalf("expected links to include ugc rel, got %s", html)
	}
	if !strings.Contains(html, `loading="lazy"`) {
		t.Fatalf("expected images to lazy load, got %s", html)
	}
	if !strings.Contains(html, `decoding="async"`) {
		t.Fatalf("expected images to decode async, got %s", html)
	}
}

func TestExtractFirstImageURL(t *testing.T) {
	got := ExtractFirstImageURL("正文\n\n![first](/file/img/first.webp)\n\n![second](https://example.com/second.webp)")
	if got != "/file/img/first.webp" {
		t.Fatalf("expected first markdown image, got %q", got)
	}
}

func TestExtractFirstImageURLSkipsInlineData(t *testing.T) {
	got := ExtractFirstImageURL("![inline](data:image/png;base64,abc)\n\n![public](https://example.com/public.webp)")
	if got != "https://example.com/public.webp" {
		t.Fatalf("expected first public image, got %q", got)
	}
}
