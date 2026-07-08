package postservice

import (
	"log/slog"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/samber/lo"
)

const topicSequenceLockShards = 256

var topicSequenceLocks [topicSequenceLockShards]sync.Mutex

func CreateTopicPost(entity *posts.Entity, topicEntity topics.Entity) error {
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

func reserveTopicPostSequence(topicId uint64) (uint64, error) {
	lock := &topicSequenceLocks[topicId%topicSequenceLockShards]
	lock.Lock()
	defer lock.Unlock()

	return topics.ReservePostSequence(topicId)
}

func SyncTopicPostStats(topicEntity topics.Entity, userId uint64, isDelete bool) {
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
