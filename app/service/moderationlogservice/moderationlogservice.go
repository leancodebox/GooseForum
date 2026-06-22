package moderationlogservice

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
)

func ArticleStatusChanged(actorUserId uint64, articleId uint64, title string, blocked bool) {
	action := moderationLog.ActionArticleUnblocked
	status := "unblocked"
	if blocked {
		action = moderationLog.ActionArticleBlocked
		status = "blocked"
	}
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      action,
		SubjectType: moderationLog.SubjectArticle,
		SubjectId:   articleId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.article.statusChanged",
			Params: map[string]any{
				"title":  title,
				"status": status,
			},
		},
	})
}

type ReplySnapshot struct {
	ReplyId       uint64
	ArticleId     uint64
	ArticleTitle  string
	ReplyNo       uint64
	ReplyAuthorId uint64
	ReplyAuthor   string
	Excerpt       string
}

type ReportSnapshot struct {
	ReportId     uint64
	TargetType   string
	TargetId     uint64
	TargetURL    string
	ArticleId    uint64
	ArticleTitle string
	ReplyNo      uint64
	Reason       string
	Resolution   string
	ReporterId   uint64
	Reporter     string
	Excerpt      string
}

func ReplyStatusChanged(actorUserId uint64, snapshot ReplySnapshot, blocked bool) {
	action := moderationLog.ActionReplyUnblocked
	status := "unblocked"
	if blocked {
		action = moderationLog.ActionReplyBlocked
		status = "blocked"
	}
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      action,
		SubjectType: moderationLog.SubjectReply,
		SubjectId:   snapshot.ReplyId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.reply.statusChanged",
			Params: map[string]any{
				"articleId":     snapshot.ArticleId,
				"title":         snapshot.ArticleTitle,
				"replyNo":       snapshot.ReplyNo,
				"replyAuthorId": snapshot.ReplyAuthorId,
				"replyAuthor":   snapshot.ReplyAuthor,
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
				"articleId":  snapshot.ArticleId,
				"title":      snapshot.ArticleTitle,
				"replyNo":    snapshot.ReplyNo,
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

func CategoryModeratorAdded(actorUserId uint64, categoryId uint64, categoryName string, userId uint64, username string) {
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      moderationLog.ActionCategoryModeratorAdded,
		SubjectType: moderationLog.SubjectCategory,
		SubjectId:   categoryId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.category.moderatorAdded",
			Params: map[string]any{
				"categoryId":   categoryId,
				"categoryName": categoryName,
				"userId":       userId,
				"username":     username,
			},
		},
	})
}

func CategoryModeratorRemoved(actorUserId uint64, categoryId uint64, categoryName string, userId uint64) {
	create(moderationLog.Entity{
		ActorUserId: actorUserId,
		Action:      moderationLog.ActionCategoryModeratorRemoved,
		SubjectType: moderationLog.SubjectCategory,
		SubjectId:   categoryId,
		Payload: moderationLog.Payload{
			MessageCode: "moderation.log.category.moderatorRemoved",
			Params: map[string]any{
				"categoryId":   categoryId,
				"categoryName": categoryName,
				"userId":       userId,
			},
		},
	})
}

func create(entity moderationLog.Entity) {
	if err := moderationLog.Create(&entity); err != nil {
		slog.Error("create moderation log failed", "action", entity.Action, "subjectType", entity.SubjectType, "subjectId", entity.SubjectId, "err", err)
	}
}
