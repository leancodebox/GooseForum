package contentmoderationservice

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/sensitivewordservice"
)

func ReviewTopic(id, version uint64) {
	if !sensitivewordservice.Enabled() {
		return
	}
	topic, err := topics.GetForModeration(id, version)
	if err != nil {
		return
	}
	if topic.FirstPostId == 0 {
		return
	}
	post := posts.Get(topic.FirstPostId)
	if post.Id == 0 {
		return
	}
	titleResult := sensitivewordservice.Check(topic.Title)
	postResult := sensitivewordservice.Check(post.Content)
	passed := titleResult.Passed && postResult.Passed
	reason := titleResult.Reason
	if postResult.Reason != "" {
		if reason != "" {
			reason += "; "
		}
		reason += postResult.Reason
	}
	status := "approved"
	visible := 1
	if !passed {
		status = "rejected"
		visible = 0
	}
	now := time.Now()
	_ = topics.UpdateModeration(id, version, map[string]any{"moderation_status": status, "moderation_reason": reason, "moderated_at": &now, "status": visible})
	if topic.FirstPostId > 0 {
		_ = posts.UpdateModerationByID(topic.FirstPostId, map[string]any{"moderation_status": status, "moderation_reason": reason, "moderated_at": &now, "process_status": 1 - visible})
	}
}

func PrepareTopic(topic *topics.Entity) {
	if sensitivewordservice.Enabled() {
		topic.ModerationStatus = "pending"
		topic.ModerationVersion++
		if sensitivewordservice.Config().Mode == "after_review" {
			topic.Status = 0
		}
	}
}
func PreparePost(post *posts.Entity) {
	if sensitivewordservice.Enabled() {
		post.ModerationStatus = "pending"
		post.ModerationVersion++
		if sensitivewordservice.Config().Mode == "after_review" {
			post.ProcessStatus = 1
		}
	}
}

func ReviewPost(id, version uint64) {
	if !sensitivewordservice.Enabled() {
		return
	}
	post, err := posts.GetForModeration(id, version)
	if err != nil {
		return
	}
	result := sensitivewordservice.Check(post.Content)
	status, process := "approved", 0
	if !result.Passed {
		status, process = "rejected", 1
	}
	now := time.Now()
	_ = posts.UpdateModeration(id, version, map[string]any{"moderation_status": status, "moderation_reason": result.Reason, "moderated_at": &now, "process_status": process})
	if post.PostNo == 1 && post.TopicId > 0 {
		topic := topics.Get(post.TopicId)
		ReviewTopic(post.TopicId, topic.ModerationVersion)
	}
}

func ValidateMode(mode string) error {
	if mode != "after_review" && mode != "visible_then_review" {
		return fmt.Errorf("invalid moderation mode %q", mode)
	}
	return nil
}
