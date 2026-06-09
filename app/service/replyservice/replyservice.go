package replyservice

import (
	"log/slog"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/articlesUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/samber/lo"
)

const articleSequenceLockShards = 256

var articleSequenceLocks [articleSequenceLockShards]sync.Mutex

func CreateArticleReply(entity *reply.Entity, articleEntity articles.SmallEntity) error {
	replyNo, err := reserveArticleReplySequence(entity.ArticleId)
	if err != nil {
		return err
	}

	entity.ReplyNo = replyNo
	if err := reply.Create(entity); err != nil {
		return err
	}

	SyncArticleReplyStats(articleEntity, entity.UserId, false)
	return nil
}

func reserveArticleReplySequence(articleId uint64) (uint64, error) {
	lock := &articleSequenceLocks[articleId%articleSequenceLockShards]
	lock.Lock()
	defer lock.Unlock()

	return articles.ReserveReplySequence(articleId)
}

func SyncArticleReplyStats(articleEntity articles.SmallEntity, userId uint64, isDelete bool) {
	if isDelete {
		if err := articlesUserStat.DecrementUserReply(articleEntity.Id, userId); err != nil {
			slog.Error("failed to decrement user reply stat", "articleId", articleEntity.Id, "userId", userId, "err", err)
		}
	} else {
		if err := articlesUserStat.IncrementUserReply(articleEntity.Id, userId); err != nil {
			slog.Error("failed to increment user reply stat", "articleId", articleEntity.Id, "userId", userId, "err", err)
		}
	}

	list := articlesUserStat.SyncArticlePosters(articleEntity.Id)
	filteredList := lo.Filter(list, func(userID uint64, _ int) bool {
		return userID != articleEntity.UserId
	})
	finalList := append([]uint64{articleEntity.UserId}, filteredList...)

	pList := lo.Map(finalList, func(userID uint64, _ int) articles.Poster {
		return articles.Poster{
			UserID: userID,
		}
	})

	if isDelete {
		if err := articles.DecrementReplyFast(articleEntity.Id, pList); err != nil {
			slog.Error("failed to decrement article reply count", "articleId", articleEntity.Id, "err", err)
		}
	} else {
		if err := articles.IncrementReplyFast(articleEntity.Id, pList); err != nil {
			slog.Error("failed to increment article reply count", "articleId", articleEntity.Id, "err", err)
		}
	}
}
