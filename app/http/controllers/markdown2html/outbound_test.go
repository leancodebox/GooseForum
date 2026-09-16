package markdown2html

import (
	"strings"
	"testing"
)

func TestOnlyPostLinksUseOutbound(t *testing.T) {
	source := "[external](https://example.com/a?q=one&b=two#part) [relative](/p/post/1) [network](//example.com/path) ![image](https://example.com/a.png)"
	post := PostMarkdownToHTML(source)
	if strings.Count(post, `href="/outbound?url=`) != 2 {
		t.Fatal(post)
	}
	if !strings.Contains(post, `href="/p/post/1"`) || !strings.Contains(post, `src="https://example.com/a.png"`) {
		t.Fatal(post)
	}
	if strings.Contains(MarkdownToHTML(source), "/outbound?") {
		t.Fatal("announcement rendering changed")
	}
}
