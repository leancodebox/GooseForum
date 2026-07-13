package forum

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/resource"
)

// TestServerTemplatesParse ensures every server-rendered template parses with
// the shared FuncMap (notably the "t" translator and sprig's "dict").
func TestServerTemplatesParse(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}
	for _, name := range []string{"home.gohtml", "search.gohtml", "user.gohtml", "topic.gohtml", "links.gohtml", "sponsors.gohtml"} {
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
	type author struct {
		ID       uint64
		Username string
	}
	type topic struct {
		URL, Title, Description string
		Categories              []category
		Author                  author
		ReplyCount, ViewCount   int
		ActivityText            string
		LastUpdateTime          string
	}

	// Populated list -> "replies"/"views" labels.
	var buf bytes.Buffer
	data := map[string]any{
		"Lang":   "en",
		"Topics": []topic{{URL: "/t/1", Title: "Hello", Description: "Topic excerpt", Author: author{ID: 1, Username: "author"}, Categories: []category{{URL: "/c/general/1", Name: "General"}}, ReplyCount: 3, ViewCount: 9, ActivityText: "now"}},
	}
	if err := tmpl.ExecuteTemplate(&buf, "partials/topic_list.gohtml", data); err != nil {
		t.Fatalf("render populated partial: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "replies") || !strings.Contains(out, "views") {
		t.Errorf("partial not localized to English: %q", out)
	}
	if !strings.Contains(out, `<table class="gf-crawler-topic-list">`) || !strings.Contains(out, "<thead>") || !strings.Contains(out, "<tbody>") {
		t.Errorf("partial should render a crawlable topic table: %q", out)
	}
	for _, want := range []string{`class="gf-crawler-topic-list"`, `class="gf-crawler-topic-main"`, `class="gf-crawler-topic-title"`, `class="gf-crawler-topic-meta"`, `class="gf-crawler-topic-excerpt"`, `class="gf-crawler-topic-number"`, `class="gf-crawler-topic-activity"`} {
		if !strings.Contains(out, want) {
			t.Errorf("partial missing crawlable list class %s: %q", want, out)
		}
	}
	if !strings.Contains(out, "Activity") {
		t.Errorf("partial missing activity column: %q", out)
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

func TestTopicTemplateRendersPostStreamNoscript(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	payload := PagePayload{
		Props: TopicDetailProps{
			Topic: TopicDetailPayload{
				ID:          10,
				Title:       "Topic title",
				Description: "Topic excerpt",
				Author:      TopicAuthorPayload{ID: 1, Username: "author"},
				CreatedAt:   "2026-07-08 10:00:00",
				UpdatedAt:   "2026-07-08 10:00:00",
			},
			PostStream: PostWindowPayload{
				Posts: []PostPayload{
					{ID: 100, PostNo: 1, RenderedContent: "<p>first post</p>", Content: "first post", Author: TopicAuthorPayload{ID: 1, Username: "author"}, CreatedAt: "2026-07-08 10:00:00"},
					{ID: 101, PostNo: 2, RenderedContent: "<p><strong>reply body</strong></p>", Content: "**reply body**", Author: TopicAuthorPayload{ID: 2, Username: "replyer"}, CreatedAt: "2026-07-08 10:01:00"},
				},
			},
		},
		Meta: PageMeta{OpenGraph: &OpenGraphMeta{}},
	}

	var buf bytes.Buffer
	if err := reg.render(&buf, "topic.gohtml", templateData{Payload: payload, Lang: "en"}); err != nil {
		t.Fatalf("render topic template: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<p>first post</p>") || !strings.Contains(out, "reply body") {
		t.Fatalf("topic noscript did not render post stream content: %s", out)
	}
	if !strings.Contains(out, "<strong>reply body</strong>") {
		t.Fatalf("topic noscript did not render reply HTML: %s", out)
	}
	for _, want := range []string{`class="gf-crawler-topic-page"`, `class="gf-crawler-post-list"`, `class="gf-crawler-post"`, `class="gf-crawler-post-meta"`, `class="gf-crawler-post-body"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("topic noscript missing crawler detail class %s: %s", want, out)
		}
	}
	if !strings.Contains(out, `href="#post-100">#1</a>`) || !strings.Contains(out, `href="#post-101">#2</a>`) {
		t.Fatalf("topic noscript should render first post and replies as posts with post numbers: %s", out)
	}
	if !strings.Contains(out, "gf-crawler-post-body img") || !strings.Contains(out, "max-height: 560px") {
		t.Fatalf("topic noscript should constrain post images in crawler styles: %s", out)
	}
}

func TestTopicTemplateRendersWithoutPostStream(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	payload := PagePayload{
		Props: TopicDetailProps{
			Topic: TopicDetailPayload{
				ID:          10,
				Title:       "Topic title",
				Description: "Topic excerpt",
				Author:      TopicAuthorPayload{ID: 1, Username: "author"},
				CreatedAt:   "2026-07-08 10:00:00",
				UpdatedAt:   "2026-07-08 10:00:00",
			},
			PostStream: PostWindowPayload{},
		},
		Meta: PageMeta{OpenGraph: &OpenGraphMeta{}},
	}

	var buf bytes.Buffer
	if err := reg.render(&buf, "topic.gohtml", templateData{Payload: payload, Lang: "en"}); err != nil {
		t.Fatalf("render topic template without post stream: %v", err)
	}
	if out := buf.String(); !strings.Contains(out, "Topic title") {
		t.Fatalf("rendered topic without post stream missing title: %s", out)
	}
}

func TestNoscriptTemplatesRenderRepresentativePayloads(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	topic := TopicPayload{
		ID:           10,
		Title:        "Topic title",
		Description:  "Topic excerpt",
		URL:          "/p/post/10",
		Categories:   []TopicCategoryPayload{{ID: 3, Name: "General", URL: "/c/general/3"}},
		ReplyCount:   2,
		ViewCount:    9,
		ActivityText: "now",
	}
	pagination := PaginationPayload{Page: 1, NextPage: 2, HasNext: true, NextURL: "/?page=2"}
	layout := LayoutPayload{Site: SitePayload{Name: "GooseForum"}}

	cases := []struct {
		name     string
		template string
		payload  PagePayload
		want     string
	}{
		{
			name:     "home",
			template: "home.gohtml",
			payload:  PagePayload{Layout: layout, Props: HomeProps{Topics: []TopicPayload{topic}, Pagination: pagination}, Meta: PageMeta{Description: "Forum description"}},
			want:     "Topic title",
		},
		{
			name:     "category",
			template: "category.gohtml",
			payload: PagePayload{Props: CategoryPageProps{
				Category:   CategoryHeaderPayload{ID: 3, Name: "General", Description: "General discussion"},
				Topics:     []TopicPayload{topic},
				Pagination: pagination,
			}},
			want: "General discussion",
		},
		{
			name:     "search",
			template: "search.gohtml",
			payload: PagePayload{Props: SearchPageProps{
				Query:      "topic",
				Total:      1,
				Topics:     []TopicPayload{topic},
				Pagination: pagination,
			}},
			want: "Topic title",
		},
		{
			name:     "links",
			template: "links.gohtml",
			payload: PagePayload{Props: LinksPageProps{
				TotalCount: 1,
				Groups: []LinkGroupPayload{{
					Name:  "Friends",
					Links: []FriendLinkPayload{{Name: "Example", Desc: "Example site", URL: "https://example.com"}},
				}},
			}},
			want: "Example site",
		},
		{
			name:     "sponsors",
			template: "sponsors.gohtml",
			payload: PagePayload{Props: SponsorsPageProps{Sections: []SponsorSectionPayload{{
				Label:    "Gold",
				Sponsors: []SponsorPayload{{Name: "Sponsor", Message: "Thanks", Link: "https://example.com"}},
			}}}},
			want: "Sponsor",
		},
		{
			name:     "user",
			template: "user.gohtml",
			payload: PagePayload{Props: UserProfileProps{
				User:       &vo.UserCard{UserId: 1, Username: "author", Nickname: "Author", AvatarUrl: "/avatar.png", Bio: "Bio", TopicCount: 1},
				Topics:     []TopicPayload{topic},
				Activities: []UserActivityPayload{{Label: "Posted", URL: "/p/post/10", ContentPreview: "Preview", CreatedAt: "2026-07-08 10:00:00"}},
			}},
			want: "Preview",
		},
		{
			name:     "error",
			template: "error.gohtml",
			payload:  PagePayload{Props: ErrorPageProps{Code: "404", Title: "Not found", MessageCode: "route.notFound"}},
			want:     "Not found",
		},
		{
			name:     "topic",
			template: "topic.gohtml",
			payload: PagePayload{
				Props: TopicDetailProps{
					Topic: TopicDetailPayload{ID: 10, Title: "Topic title", Description: "Topic excerpt", Author: TopicAuthorPayload{ID: 1, Username: "author"}, CreatedAt: "2026-07-08 10:00:00", UpdatedAt: "2026-07-08 10:00:00"},
					PostStream: PostWindowPayload{Posts: []PostPayload{
						{ID: 100, PostNo: 1, RenderedContent: "<p>first post</p>", Content: "first post", Author: TopicAuthorPayload{ID: 1, Username: "author"}, CreatedAt: "2026-07-08 10:00:00"},
						{ID: 101, PostNo: 2, Content: "reply body", Author: TopicAuthorPayload{ID: 2, Username: "replyer"}, CreatedAt: "2026-07-08 10:01:00"},
					}},
				},
				Meta: PageMeta{OpenGraph: &OpenGraphMeta{}},
			},
			want: "reply body",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := reg.render(&buf, tt.template, templateData{Payload: tt.payload, Lang: "en"}); err != nil {
				t.Fatalf("render %s: %v", tt.template, err)
			}
			if out := buf.String(); !strings.Contains(out, tt.want) {
				t.Fatalf("rendered %s missing %q: %s", tt.template, tt.want, out)
			}
		})
	}
}

func TestUserTemplateRendersCrawlerProfileStructure(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	payload := PagePayload{
		Props: UserProfileProps{
			User: &vo.UserCard{
				UserId:            1,
				Username:          "author",
				Nickname:          "Author",
				AvatarUrl:         "/avatar.png",
				Bio:               "User bio",
				TopicCount:        2,
				ReplyCount:        3,
				LikeReceivedCount: 4,
				FollowerCount:     5,
			},
			Topics:     []TopicPayload{{ID: 10, Title: "Topic title", URL: "/p/post/10", Description: "Topic excerpt"}},
			Activities: []UserActivityPayload{{Label: "Posted", URL: "/p/post/10", ContentPreview: "Preview", CreatedAt: "2026-07-08 10:00:00"}},
		},
	}

	var buf bytes.Buffer
	if err := reg.render(&buf, "user.gohtml", templateData{Payload: payload, Lang: "en"}); err != nil {
		t.Fatalf("render user template: %v", err)
	}
	out := buf.String()
	for _, want := range []string{`class="gf-crawler-user-page"`, `class="gf-crawler-user-head"`, `class="gf-crawler-user-avatar"`, `class="gf-crawler-user-meta"`, `class="gf-crawler-user-stats"`, `class="gf-crawler-activity-list"`, `class="gf-crawler-activity"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("user noscript missing crawler profile class %s: %s", want, out)
		}
	}
	if !strings.Contains(out, "Author") || !strings.Contains(out, "@author") || !strings.Contains(out, "Preview") {
		t.Fatalf("user noscript missing profile content: %s", out)
	}
}

func TestListTemplatesRenderMetaPaginationLinks(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	topic := TopicPayload{ID: 10, Title: "Topic title", URL: "/p/post/10"}
	pagination := PaginationPayload{Page: 2, NextPage: 3, HasNext: true, NextURL: "/?page=3"}
	meta := PageMeta{PrevURL: "/?page=1", NextURL: "/?page=3"}

	cases := []struct {
		name     string
		template string
		payload  PagePayload
	}{
		{
			name:     "home",
			template: "home.gohtml",
			payload:  PagePayload{Props: HomeProps{Topics: []TopicPayload{topic}, Pagination: pagination}, Meta: meta},
		},
		{
			name:     "category",
			template: "category.gohtml",
			payload: PagePayload{Props: CategoryPageProps{
				Category:   CategoryHeaderPayload{ID: 3, Name: "General"},
				Topics:     []TopicPayload{topic},
				Pagination: pagination,
			}, Meta: meta},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := reg.render(&buf, tt.template, templateData{Payload: tt.payload, Lang: "en"}); err != nil {
				t.Fatalf("render %s: %v", tt.template, err)
			}
			out := buf.String()
			if !strings.Contains(out, `href="/?page=1" rel="prev"`) || !strings.Contains(out, `href="/?page=3" rel="next"`) {
				t.Fatalf("rendered %s missing prev/next no-js links: %s", tt.template, out)
			}
		})
	}
}

func TestNoscriptLayoutRendersFooterHint(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	payload := PagePayload{
		Layout: LayoutPayload{Site: SitePayload{Name: "GooseForum"}},
		Props:  HomeProps{},
	}
	var buf bytes.Buffer
	if err := reg.render(&buf, "home.gohtml", templateData{Payload: payload, Lang: "en"}); err != nil {
		t.Fatalf("render home template: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="gf-crawler-footer"`) {
		t.Fatalf("noscript layout missing footer: %s", out)
	}
	if !strings.Contains(out, `<style data-goose-crawler-style>`) {
		t.Fatalf("noscript layout missing crawler style inside noscript: %s", out)
	}
	if !strings.Contains(out, "@media (max-width: 700px)") || !strings.Contains(out, "grid-template-columns: 1fr") {
		t.Fatalf("noscript layout missing mobile crawler list styles: %s", out)
	}
	if strings.Index(out, `<style data-goose-crawler-style>`) < strings.Index(out, `<noscript>`) {
		t.Fatalf("crawler style should render inside noscript, not in head: %s", out)
	}
	if !strings.Contains(out, `href="/"`) || !strings.Contains(out, "Back to home") {
		t.Fatalf("noscript footer missing home link: %s", out)
	}
	if !strings.Contains(out, "Enable JavaScript for the best experience") {
		t.Fatalf("noscript footer missing localized JavaScript hint: %s", out)
	}
}

func TestNoscriptLayoutRendersHeaderHomeLink(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	payload := PagePayload{
		Layout: LayoutPayload{Site: SitePayload{Name: "GooseForum"}},
		Props:  HomeProps{},
	}
	var buf bytes.Buffer
	if err := reg.render(&buf, "home.gohtml", templateData{Payload: payload, Lang: "en"}); err != nil {
		t.Fatalf("render home template: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `class="gf-crawler-header"`) {
		t.Fatalf("noscript layout missing header: %s", out)
	}
	if !strings.Contains(out, `<a class="gf-crawler-home-link" href="/">GooseForum</a>`) {
		t.Fatalf("noscript header missing home link: %s", out)
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
