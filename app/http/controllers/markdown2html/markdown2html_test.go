package markdown2html

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

type markdownCompatCase struct {
	Name        string   `json:"name"`
	Markdown    string   `json:"markdown"`
	Contains    []string `json:"contains"`
	NotContains []string `json:"notContains"`
}

func TestMarkdownVersions(t *testing.T) {
	if got := GetVersion(); got != 3 {
		t.Fatalf("GetVersion() = %d, want 3", got)
	}
	if got := GetCommentVersion(); got != 1 {
		t.Fatalf("GetCommentVersion() = %d, want 1", got)
	}
	if GetParser() == nil {
		t.Fatal("GetParser() returned nil")
	}
}

func TestMarkdownToHTMLRendersGFM(t *testing.T) {
	html := MarkdownToHTML("# Title\n\n- [x] done\n\n| A | B |\n|---|---|\n| 1 | 2 |")

	for _, want := range []string{`<h1 id="title">Title</h1>`, `type="checkbox"`, `<table>`} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered HTML to contain %q, got %s", want, html)
		}
	}
}

func TestMarkdownCompatibilityFixtures(t *testing.T) {
	for _, fixture := range loadMarkdownCompatFixtures(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			html := MarkdownToHTML(fixture.Markdown)
			for _, want := range fixture.Contains {
				if !strings.Contains(html, want) {
					t.Fatalf("expected rendered HTML to contain %q, got %s", want, html)
				}
			}
			for _, unwanted := range fixture.NotContains {
				if strings.Contains(html, unwanted) {
					t.Fatalf("expected rendered HTML not to contain %q, got %s", unwanted, html)
				}
			}
		})
	}
}

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

func TestNormalizeCommentHTMLUpdatesExistingAttributes(t *testing.T) {
	html := normalizeCommentHTML(`<p><a href="https://example.com" target="_self" rel="next">link</a><img src="/a.png" loading="eager"></p>`)

	for _, want := range []string{`target="_blank"`, `rel="nofollow ugc noopener noreferrer"`, `loading="lazy"`, `decoding="async"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected normalized HTML to contain %q, got %s", want, html)
		}
	}
	if strings.Contains(html, `target="_self"`) || strings.Contains(html, `rel="next"`) || strings.Contains(html, `loading="eager"`) {
		t.Fatalf("expected existing attributes to be replaced, got %s", html)
	}
}

func TestExtractFirstImageURL(t *testing.T) {
	got := ExtractFirstImageURL("正文\n\n![first](/file/img/first.webp)\n\n![second](https://example.com/second.webp)")
	if got != "/file/img/first.webp" {
		t.Fatalf("expected first markdown image, got %q", got)
	}
}

func TestExtractFirstImageURLSkipsNonPublicDestinations(t *testing.T) {
	tests := []string{
		"",
		"![blob](blob:https://example.com/id)",
		"![ftp](ftp://example.com/a.png)",
		"![bad](http://%zz)",
		"![relative](images/a.png)",
	}

	for _, input := range tests {
		if got := ExtractFirstImageURL(input); got != "" {
			t.Fatalf("ExtractFirstImageURL(%q) = %q, want empty", input, got)
		}
	}
}

func TestExtractDescriptionSkipsCodeAndImages(t *testing.T) {
	got := ExtractDescription(`# Heading

短

This is a useful paragraph.

![logo](/logo.png)

`+"```go\nfmt.Println(\"hidden\")\n```"+`

- Another useful point.`, 200)

	if !strings.Contains(got, "Heading") || !strings.Contains(got, "This is a useful paragraph") || !strings.Contains(got, "Another useful point") {
		t.Fatalf("description missed readable text: %q", got)
	}
	if strings.Contains(got, "hidden") || strings.Contains(got, "logo") || strings.Contains(got, "短") {
		t.Fatalf("description included skipped content: %q", got)
	}
}

func TestExtractDescriptionUsesDefaultAndTruncatesRunes(t *testing.T) {
	input := strings.Repeat("鹅", 210)
	got := ExtractDescription(input, 0)

	if runeCount := len([]rune(strings.TrimSuffix(got, "..."))); runeCount != 200 {
		t.Fatalf("description rune count = %d, want 200", runeCount)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated description to end with ellipsis, got %q", got)
	}
}

func TestExtractDescriptionRt(t *testing.T) {
	input := "rt,xxxx121"
	got := ExtractDescription(input, 0)

	if got != input {
		t.Fatalf("ExtractDescription() = %q, want %q", got, input)
	}
}

func TestFallbackExtractDescription(t *testing.T) {
	got := fallbackExtractDescription(`# Title

- Useful list item with enough text
![skip](/image.png)
`+"```go\nfmt.Println(\"skip\")\n```"+`
Plain paragraph with enough words`, 18)

	if got != "Useful list item w..." {
		t.Fatalf("fallbackExtractDescription() = %q", got)
	}
}

func TestExtractSearchContentPreservesUsefulMarkdownContext(t *testing.T) {
	got := ExtractSearchContent(`# Heading

Text with [Goose](https://gooseforum.online) and ` + "`code`" + `.

![logo](/logo.png)

---

<section>hidden html</section>

` + "```go\nhidden()\n```")

	for _, want := range []string{"Heading", "Goose](https://gooseforum.online)", "` code`", "logo](/logo.png)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("search content missing %q: %q", want, got)
		}
	}
	if strings.Contains(got, "hidden html") || strings.Contains(got, "hidden()") {
		t.Fatalf("search content included skipped content: %q", got)
	}
}

func TestCompactWhitespace(t *testing.T) {
	got := compactWhitespace("  first  \n\n second\t\n")
	if got != "first\nsecond" {
		t.Fatalf("compactWhitespace() = %q", got)
	}
}

func TestFindFirstElementAndShouldInsertSpaceFallbacks(t *testing.T) {
	if got := findFirstElement(nil, "div"); got != nil {
		t.Fatalf("findFirstElement(nil) = %#v, want nil", got)
	}

	root := &ast.Document{}
	if shouldInsertSpace([]ast.Node{root}) {
		t.Fatal("shouldInsertSpace() should be false outside paragraph/heading")
	}
}

func TestExtractFirstImageURLSkipsInlineData(t *testing.T) {
	got := ExtractFirstImageURL("![inline](data:image/png;base64,abc)\n\n![public](https://example.com/public.webp)")
	if got != "https://example.com/public.webp" {
		t.Fatalf("expected first public image, got %q", got)
	}
}

func loadMarkdownCompatFixtures(t *testing.T) []markdownCompatCase {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to locate markdown test file")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "testdata", "markdown-compat")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}

	fixtures := make([]markdownCompatCase, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		markdown, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		expect, err := os.ReadFile(filepath.Join(root, name+".json"))
		if err != nil {
			t.Fatal(err)
		}

		fixture := markdownCompatCase{Name: name, Markdown: string(markdown)}
		if err := json.Unmarshal(expect, &fixture); err != nil {
			t.Fatal(err)
		}
		fixtures = append(fixtures, fixture)
	}
	sort.Slice(fixtures, func(i, j int) bool { return fixtures[i].Name < fixtures[j].Name })
	return fixtures
}
