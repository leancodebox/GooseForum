package replyservice

import (
	"log/slog"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/articlesUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/samber/lo"
)

const articleSequenceLockShards = 256

var articleSequenceLocks [articleSequenceLockShards]sync.Mutex

func CreateTopicPost(entity *posts.Entity, topicEntity topics.SmallEntity) error {
	postNo, err := reserveTopicPostSequence(entity.TopicId)
	if err != nil {
		return err
	}

	entity.PostNo = postNo
	if err := posts.Create(entity); err != nil {
		return err
	}

	SyncTopicPostStats(topicEntity, entity.UserId, false)
	return nil
}

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

func reserveTopicPostSequence(topicId uint64) (uint64, error) {
	lock := &articleSequenceLocks[topicId%articleSequenceLockShards]
	lock.Lock()
	defer lock.Unlock()

	return topics.ReservePostSequence(topicId)
}

func SyncTopicPostStats(topicEntity topics.SmallEntity, userId uint64, isDelete bool) {
	if isDelete {
		if err := topicUserStat.DecrementUserPost(topicEntity.Id, userId); err != nil {
			slog.Error("failed to decrement topic user post stat", "topicId", topicEntity.Id, "userId", userId, "err", err)
		}
	} else {
		if err := topicUserStat.IncrementUserPost(topicEntity.Id, userId); err != nil {
			slog.Error("failed to increment topic user post stat", "topicId", topicEntity.Id, "userId", userId, "err", err)
		}
	}

	list := topicUserStat.SyncTopicPosters(topicEntity.Id)
	filteredList := lo.Filter(list, func(userID uint64, _ int) bool {
		return userID != topicEntity.UserId
	})
	finalList := append([]uint64{topicEntity.UserId}, filteredList...)

	pList := lo.Map(finalList, func(userID uint64, _ int) topics.Poster {
		return topics.Poster{
			UserID: userID,
		}
	})

	if isDelete {
		if err := topics.DecrementPostFast(topicEntity.Id, pList); err != nil {
			slog.Error("failed to decrement topic post count", "topicId", topicEntity.Id, "err", err)
		}
	} else {
		if err := topics.IncrementPostFast(topicEntity.Id, pList); err != nil {
			slog.Error("failed to increment topic post count", "topicId", topicEntity.Id, "err", err)
		}
	}
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
