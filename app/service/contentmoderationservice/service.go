package contentmoderationservice

import (
	"fmt"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
)

// ReviewTopic only changes the supplied write snapshot. Call before opening a
// transaction, and persist the result through the normal topic write path.
func ReviewTopic(topic *topics.Entity, post *posts.Entity) {
	previouslyRejected := post.ModerationStatus == "rejected" || post.ModerationStatus == "pending" || post.ModerationStatus == "denied"
	topic.ModerationVersion++
	post.ModerationVersion++
	topic.ModerationStatus, topic.ModerationReason, topic.ModeratedAt = "none", "", nil
	post.ModerationStatus, post.ModerationReason, post.ModeratedAt = "none", "", nil
	// Drafts are private and must never be published by moderation.
	if topic.Status == 1 && sensitivewordservice.Enabled() {
		title := sensitivewordservice.Check(topic.Title)
		body := sensitivewordservice.Check(post.Content)
		topic.Title, post.Content = title.Content, body.Content
		topic.ModerationStatus = "approved"
		if !title.Passed || !body.Passed {
			topic.ModerationStatus = "rejected"
			topic.Status = 0
		}
		now := time.Now()
		topic.ModeratedAt = &now
		topic.ModerationReason = boundedReason(strings.Trim(strings.Join([]string{title.Reason, body.Reason}, "; "), "; "))
		post.ModerationStatus, post.ModerationReason, post.ModeratedAt = topic.ModerationStatus, topic.ModerationReason, &now
	}
	// First-post visibility follows its topic, not an independent auto-review flag.
	// Clear only the legacy automatic rejection flag; retain manual topic blocks.
	if post.ProcessStatus == 1 && previouslyRejected {
		post.ProcessStatus = 0
	}
	topic.Excerpt = markdown2html.ExtractDescription(post.Content, 200)
	topic.FirstImageURL = markdown2html.ExtractFirstImageURL(post.Content)
	post.RenderedHTML = markdown2html.PostMarkdownToHTML(post.Content)
	post.RenderedVersion = markdown2html.GetPostVersion()
}

func ReviewPost(post *posts.Entity) {
	if post.WasPublished() && post.PublishedAt == nil {
		publishedAt := post.CreatedAt
		post.PublishedAt = &publishedAt
	}
	previouslyRejected := post.ModerationStatus == "rejected" || post.ModerationStatus == "pending" || post.ModerationStatus == "denied"
	post.ModerationVersion++
	post.ModerationStatus, post.ModerationReason, post.ModeratedAt = "none", "", nil
	if previouslyRejected {
		post.ProcessStatus = 0
	}
	if sensitivewordservice.Enabled() {
		result := sensitivewordservice.Check(post.Content)
		post.Content = result.Content
		post.ModerationStatus = "approved"
		if !result.Passed {
			post.ModerationStatus, post.ProcessStatus = "rejected", 1
		}
		now := time.Now()
		post.ModeratedAt = &now
		post.ModerationReason = boundedReason(result.Reason)
	}
	if post.ProcessStatus == 0 && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}
	post.RenderedHTML = markdown2html.PostMarkdownToHTML(post.Content)
	post.RenderedVersion = markdown2html.GetPostVersion()
}

func boundedReason(reason string) string {
	runes := []rune(reason)
	if len(runes) > 512 {
		return string(runes[:509]) + "..."
	}
	return reason
}

// Retained for clients using the original settings payload. Detection is now
// synchronous in both modes, before any public side effects.
func ValidateMode(mode string) error {
	if mode != "after_review" && mode != "visible_then_review" {
		return fmt.Errorf("invalid moderation mode %q", mode)
	}
	return nil
}
