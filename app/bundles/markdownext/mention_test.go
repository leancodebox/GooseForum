package markdownext

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

func TestMentionContexts(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Linkify, Extension{}))
	for _, test := range []struct {
		source string
		count  int
		notify bool
	}{
		{"你好 @alice", 1, true}, {"**@alice**", 1, true}, {"`@alice`", 0, false}, {"```\n@alice\n```", 0, false}, {"\\@alice", 0, false}, {"mail@alice.com", 0, false}, {"https://example.com/@alice", 0, false}, {"[text @alice](https://example.com)", 0, false}, {"![text @alice](/image)", 0, false}, {"> @alice", 1, false}, {"# @alice", 1, false}, {"[mention user=\"123\"]@alice[/mention]", 1, true}, {"[text [mention user=\"123\"]@alice[/mention]](/url)", 0, false}, {"<span>@alice</span>", 0, false},
	} {
		t.Run(test.source, func(t *testing.T) {
			doc := md.Parser().Parse(text.NewReader([]byte(test.source)))
			count := 0
			_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					if node, ok := n.(*Node); ok {
						count++
						if NotificationsAllowed(node) != test.notify {
							t.Errorf("wrong notification context")
						}
					}
				}
				return ast.WalkContinue, nil
			})
			if count != test.count {
				t.Errorf("got %d nodes, want %d", count, test.count)
			}
		})
	}
}
func TestTagGrammar(t *testing.T) {
	for _, source := range []string{"[note kind=\"warning\"]", "[note text=\"a\\\"b\\]c\\\\d\"]"} {
		if _, _, _, ok := ParseTag(source); !ok {
			t.Errorf("rejected %s", source)
		}
	}
	for _, source := range []string{"[Note]", "[note kind='warning']", "[note kind=\"a\" kind=\"b\"]", "[note kind=\"a\nb\"]", "[note value=\"" + strings.Repeat("a", 1100) + "\"]"} {
		if _, _, _, ok := ParseTag(source); ok {
			t.Errorf("accepted %s", source)
		}
	}
}

func TestOrdinaryMarkdownAndUnknownExtensions(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Linkify, Extension{}))
	for _, test := range []struct{ source, want string }{
		{"[test](/url)", `href="/url"`},
		{"[mention](/url)", `href="/url"`},
		{"[ref]\n\n[ref]: /url", `href="/url"`},
		{"1 < 2 @alice.", "@alice."},
	} {
		var output strings.Builder
		if err := md.Convert([]byte(test.source), &output); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), test.want) {
			t.Fatalf("%q rendered %s", test.source, output.String())
		}
	}
	for _, source := range []string{"[future]\n@alice\n[/future]", "[future][mention user=\"123\"]@alice[/mention][/future]"} {
		doc := md.Parser().Parse(text.NewReader([]byte(source)))
		_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if node, ok := n.(*Node); ok && entering && (node.UserID != 0 || node.Username != "") {
				t.Errorf("parsed mention inside unknown extension")
			}
			return ast.WalkContinue, nil
		})
	}
}

func TestOpaqueBlockBoundaries(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Linkify, Extension{}))
	for _, source := range []string{
		"[future]\n\n**hidden**\n\n![hidden](/hidden.webp)\n\n[/future]\n\n**visible**\n\n![visible](/visible.webp)",
		"> [future]\n>\n> **hidden**\n>\n> ![hidden](/hidden.webp)\n>\n> [/future]\n\n**visible**\n\n![visible](/visible.webp)",
	} {
		var output strings.Builder
		if err := md.Convert([]byte(source), &output); err != nil {
			t.Fatal(err)
		}
		html := output.String()
		if strings.Contains(html, "<strong>hidden") || strings.Contains(html, `src="/hidden.webp"`) || !strings.Contains(html, "<strong>visible</strong>") || !strings.Contains(html, `src="/visible.webp"`) {
			t.Fatalf("opaque block crossed its boundary: %s", html)
		}
	}
	var output strings.Builder
	if err := md.Convert([]byte("[future]\n\n**hidden**\n\n[/future]"), &output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "<strong>") || !strings.Contains(output.String(), "[/future]") {
		t.Fatalf("opaque EOF delimiter lost: %s", output.String())
	}
}
