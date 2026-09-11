package contentmoderationservice

import (
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
)

// PrepareTopic records publication intent before saving. Only published content
// enters the queue; changing a topic to a draft invalidates older review jobs.
func PrepareTopic(topic *topics.Entity, post *posts.Entity) {
	config := sensitivewordservice.Config()
	previouslyHidden := post.ModerationStatus == "pending" || post.ModerationStatus == "rejected" || post.ModerationStatus == "denied"
	topic.ModerationVersion++
	post.ModerationVersion++
	topic.ModerationStatus, topic.ModerationReason, topic.ModeratedAt = "none", "", nil
	if topic.Status == 1 && config.Enabled {
		topic.ModerationStatus = "pending"
		if config.Mode == pageConfig.ModerationAfterReview {
			topic.Status = 0
		}
	}
	post.ModerationStatus, post.ModerationReason, post.ModeratedAt = topic.ModerationStatus, "", nil
	if previouslyHidden {
		post.ProcessStatus = 0
	}
	topic.Excerpt = markdown2html.ExtractDescription(post.Content, 200)
	topic.FirstImageURL = markdown2html.ExtractFirstImageURL(post.Content)
	post.RenderedHTML = markdown2html.PostMarkdownToHTML(post.Content)
	post.RenderedVersion = markdown2html.GetPostVersion()
}

func PreparePost(post *posts.Entity) {
	config := sensitivewordservice.Config()
	if post.WasPublished() && post.PublishedAt == nil {
		publishedAt := post.CreatedAt
		post.PublishedAt = &publishedAt
	}
	previouslyHidden := post.ModerationStatus == "pending" || post.ModerationStatus == "rejected" || post.ModerationStatus == "denied"
	manuallyBlocked := post.ProcessStatus == 1 && !previouslyHidden
	post.ModerationVersion++
	post.ModerationStatus, post.ModerationReason, post.ModeratedAt = "none", "", nil
	if previouslyHidden {
		post.ProcessStatus = 0
	}
	if config.Enabled && !manuallyBlocked {
		post.ModerationStatus = "pending"
		if config.Mode == pageConfig.ModerationAfterReview {
			post.ProcessStatus = 1
		}
	}
	if post.ProcessStatus == 0 && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}
	post.RenderedHTML = markdown2html.PostMarkdownToHTML(post.Content)
	post.RenderedVersion = markdown2html.GetPostVersion()
}
