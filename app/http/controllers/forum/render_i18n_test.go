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
	type topic struct {
		URL, Title, Description string
		Categories              []category
		ReplyCount, ViewCount   int
		ActivityText            string
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
					{ID: 101, PostNo: 2, Content: "reply body", Author: TopicAuthorPayload{ID: 2, Username: "replyer"}, CreatedAt: "2026-07-08 10:01:00"},
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
}

func TestNoscriptTemplatesRenderRepresentativePayloads(t *testing.T) {
	reg, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		t.Fatalf("newRegistry: %v", err)
	}

	topic := TopicPayload{
		ID:          10,
		Title:       "Topic title",
		Description: "Topic excerpt",
		URL:         "/p/post/10",
		Categories:  []TopicCategoryPayload{{ID: 3, Name: "General", URL: "/c/general/3"}},
		ReplyCount:  2,
		ViewCount:   9,
		ActivityText:"now",
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

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
