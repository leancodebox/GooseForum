package postservice

import (
	"log/slog"
	"sync"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

const topicSequenceLockShards = 256

var topicSequenceLocks [topicSequenceLockShards]sync.Mutex

func CreateTopicPost(entity *posts.Entity, topicEntity topics.Entity) error {
	lock := &topicSequenceLocks[entity.TopicId%topicSequenceLockShards]
	lock.Lock()
	defer lock.Unlock()

	return dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		postNo, err := topics.ReservePostSequenceWithDB(tx, entity.TopicId)
		if err != nil {
			return err
		}
		entity.PostNo = postNo
		if err := posts.CreateWithDB(tx, entity); err != nil {
			return err
		}
		if err := topicUserStat.IncrementUserPostWithDB(tx, topicEntity.Id, entity.UserId); err != nil {
			return err
		}
		activeUserIDs, err := topicUserStat.SyncTopicPostersWithDB(tx, topicEntity.Id)
		if err != nil {
			return err
		}
		posters := buildPosters(topicEntity.UserId, activeUserIDs)
		return topics.IncrementPostFastWithDB(tx, topicEntity.Id, posters, entity.Id, entity.CreatedAt)
	})
}

func SyncTopicPostStats(topicEntity topics.Entity, postEntity posts.Entity, isDelete bool) {
	userId := postEntity.UserId
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
	pList := buildPosters(topicEntity.UserId, list)

	if isDelete {
		lastPost, _ := posts.GetLastByTopicID(topicEntity.Id)
		if err := topics.DecrementPostFast(topicEntity.Id, pList, lastPost.Id, lastPost.CreatedAt); err != nil {
			slog.Error("failed to decrement topic post count", "topicId", topicEntity.Id, "err", err)
		}
	} else {
		if err := topics.IncrementPostFast(topicEntity.Id, pList, postEntity.Id, postEntity.CreatedAt); err != nil {
			slog.Error("failed to increment topic post count", "topicId", topicEntity.Id, "err", err)
		}
	}
}

func buildPosters(topicAuthorID uint64, activeUserIDs []uint64) []topics.Poster {
	list := activeUserIDs
	filteredList := lo.Filter(list, func(userID uint64, _ int) bool {
		return userID != topicAuthorID
	})
	finalList := append([]uint64{topicAuthorID}, filteredList...)

	return lo.Map(finalList, func(userID uint64, _ int) topics.Poster {
		return topics.Poster{
			UserID: userID,
		}
	})
}
