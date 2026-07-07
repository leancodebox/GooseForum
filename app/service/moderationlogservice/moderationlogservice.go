package moderationlogservice

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
)

func TopicStatusChanged(actorUserId uint64, topicId uint64, title string, blocked bool) {
	action := moderationLog.ActionTopicUnblocked
	status := "unblocked"
	if blocked {
		action = moderationLog.ActionTopicBlocked
		status = "blocked"
	}
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      action,
		SubjectType: moderationLog.SubjectTopic,
		SubjectId:   topicId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.article.statusChanged",
			Params: map[string]any{
				"topicId": topicId,
				"title":   title,
				"status":  status,
			},
		},
	})
}

type PostSnapshot struct {
	PostId       uint64
	TopicId      uint64
	TopicTitle   string
	PostNo       uint64
	PostAuthorId uint64
	PostAuthor   string
	Excerpt      string
}

type ReportSnapshot struct {
	ReportId   uint64
	TargetType string
	TargetId   uint64
	TargetURL  string
	TopicId    uint64
	TopicTitle string
	PostNo     uint64
	Reason     string
	Resolution string
	ReporterId uint64
	Reporter   string
	Excerpt    string
}

func PostStatusChanged(actorUserId uint64, snapshot PostSnapshot, blocked bool) {
	action := moderationLog.ActionPostUnblocked
	status := "unblocked"
	if blocked {
		action = moderationLog.ActionPostBlocked
		status = "blocked"
	}
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      action,
		SubjectType: moderationLog.SubjectPost,
		SubjectId:   snapshot.PostId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.reply.statusChanged",
			Params: map[string]any{
				"topicId":       snapshot.TopicId,
				"postId":        snapshot.PostId,
				"title":         snapshot.TopicTitle,
				"postNo":        snapshot.PostNo,
				"replyAuthorId": snapshot.PostAuthorId,
				"replyAuthor":   snapshot.PostAuthor,
				"excerpt":       snapshot.Excerpt,
				"status":        status,
			},
		},
	})
}

func ReportStatusChanged(actorUserId uint64, snapshot ReportSnapshot, status string) {
	action := moderationLog.ActionReportResolved
	if status == "rejected" {
		action = moderationLog.ActionReportRejected
	}
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      action,
		SubjectType: moderationLog.SubjectReport,
		SubjectId:   snapshot.ReportId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.report.statusChanged",
			Params: map[string]any{
				"targetType": snapshot.TargetType,
				"targetId":   snapshot.TargetId,
				"targetUrl":  snapshot.TargetURL,
				"topicId":    snapshot.TopicId,
				"title":      snapshot.TopicTitle,
				"postNo":     snapshot.PostNo,
				"reason":     snapshot.Reason,
				"resolution": snapshot.Resolution,
				"reporterId": snapshot.ReporterId,
				"reporter":   snapshot.Reporter,
				"excerpt":    snapshot.Excerpt,
				"status":     status,
			},
		},
	})
}

func create(entity moderationLog.Entity) {
	if err := moderationLog.Create(&entity); err != nil {
		slog.Error("create moderation log failed", "action", entity.Action, "subjectType", entity.SubjectType, "subjectId", entity.SubjectId, "err", err)
	}
}
