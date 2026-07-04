package forum

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

	"github.com/leancodebox/GooseForum/resource"
)

// TestServerTemplatesParse ensures every server-rendered template parses with
// the shared FuncMap (notably the "t" translator and sprig's "dict").
func TestServerTemplatesParse(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}
	for _, name := range []string{"home.gohtml", "search.gohtml", "user.gohtml", "article.gohtml", "links.gohtml", "sponsors.gohtml"} {
		if reg.templates[name] == nil {
			t.Errorf("template %s failed to register", name)
		}
	}
}

// TestTopicListPartialLocalized renders the topic list partial through the
// dict("Topics", ..., "Lang", ...) contract and checks it localizes to English
// with no residual Chinese.
func TestTopicListPartialLocalized(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}
	tmpl := reg.templates["home.gohtml"]
	if tmpl == nil {
		t.Fatal("home.gohtml missing")
	}

	type category struct{ URL, Name string }
	type topic struct {
		URL, Title, Description string
		Categories             []category
		ReplyCount, ViewCount  int
		ActivityText           string
	}

	// Populated list -> "replies"/"views" labels.
	var buf bytes.Buffer
	data := map[string]any{
		"Lang":   "en",
		"Topics": []topic{{URL: "/t/1", Title: "Hello", ReplyCount: 3, ViewCount: 9, ActivityText: "now"}},
	}
	if err := tmpl.ExecuteTemplate(&buf, "partials/topic_list.gohtml", data); err != nil {
		t.Fatalf("render populated partial: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "replies") || !strings.Contains(out, "views") {
		t.Errorf("partial not localized to English: %q", out)
	}
	if hasHan(out) {
		t.Errorf("partial still contains Chinese: %q", out)
	}

	// Empty list -> localized "No topics yet." message.
	buf.Reset()
	if err := tmpl.ExecuteTemplate(&buf, "partials/topic_list.gohtml", map[string]any{"Lang": "en", "Topics": []topic{}}); err != nil {
		t.Fatalf("render empty partial: %v", err)
	}
	if empty := buf.String(); !strings.Contains(empty, "No topics yet") {
		t.Errorf("empty partial not localized: %q", empty)
	}
}

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
