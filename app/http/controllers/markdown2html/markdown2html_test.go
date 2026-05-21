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
