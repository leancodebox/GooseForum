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
