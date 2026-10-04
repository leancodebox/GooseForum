package mentionservice

import (
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/markdownext"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/unreadservice"
)

func Recipients(post posts.Entity) []uint64 {
	if post.SourceVersion == 0 {
		return nil
	}
	set := map[uint64]bool{}
	for _, id := range post.LegacyMentionIDs {
		set[id] = true
	}
	var ids []uint64
	for _, node := range Nodes(post.Content) {
		if node.UserID > 0 && markdownext.NotificationsAllowed(node) && !set[node.UserID] {
			set[node.UserID] = true
			ids = append(ids, node.UserID)
		}
	}
	return ids
}

// Existing public text must not notify historical recipients when its syntax is upgraded.
func PreserveLegacyMentions(post *posts.Entity, nextVersion uint8, wasPublished bool) error {
	if !wasPublished || post.SourceVersion != 0 || nextVersion == 0 {
		return nil
	}
	canonical, err := Normalize(post.Content, nextVersion, false)
	if err != nil {
		return err
	}
	seen := map[uint64]bool{}
	for _, id := range post.LegacyMentionIDs {
		seen[id] = true
	}
	for _, id := range Recipients(posts.Entity{Content: canonical, SourceVersion: nextVersion}) {
		if !seen[id] {
			post.LegacyMentionIDs = append(post.LegacyMentionIDs, id)
			seen[id] = true
		}
	}
	return nil
}
func Notify(topic topics.Entity, post posts.Entity) {
	if post.SourceVersion == 0 || post.UserId == 0 || post.ProcessStatus != 0 || post.ModerationStatus == "pending" || post.ModerationStatus == "rejected" || post.ModerationStatus == "denied" || topic.Status != 1 || topic.ProcessStatus != 0 || topic.ModerationStatus == "pending" || post.DeletedAt.Valid {
		return
	}
	ids := Recipients(post)
	if len(ids) > 20 {
		return
	}
	identities, err := users.MentionIdentities(ids, nil)
	if err != nil {
		slog.Warn("load mentioned users failed", "error", err)
		return
	}
	eligible := []uint64{}
	for _, user := range identities {
		if user.Id != post.UserId && user.EffectiveRestriction(time.Now()) != users.RestrictionBanned {
			eligible = append(eligible, user.Id)
		}
	}
	readable, err := accesscontrol.FilterReadableUserIDs(eligible, topic.MainCategoryId)
	if err != nil {
		return
	}
	preview := markdown2html.ExtractPreviewVersion(post.Content, 64, post.SourceVersion)
	for _, id := range readable {
		key := DedupeKey(post.Id, id)
		notification := eventNotification.Entity{UserId: id, TopicId: topic.Id, EventType: eventNotification.EventTypeMention, DedupeKey: &key, Payload: eventNotification.NotificationPayload{TemplateKey: eventNotification.TemplateMention, TemplateParams: eventNotification.NotificationTemplateParams{Preview: preview}, Content: preview, ActorId: post.UserId, TopicId: topic.Id, PostId: post.Id, PostNo: post.PostNo}}
		if err := eventNotification.Create(&notification); err != nil {
			slog.Warn("send mention notification failed", "postId", post.Id, "error", err)
		} else {
			unreadservice.Invalidate(id)
		}
	}
}
