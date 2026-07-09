package pointservice

import (
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type PointsAction int

const (
	PointsActionUnknown PointsAction = iota
	PointsActionInit
	PointsActionTopicPublished
	PointsActionPostCreated
)

func (action PointsAction) Code() string {
	switch action {
	case PointsActionInit:
		return "init"
	case PointsActionTopicPublished:
		return "topic_published"
	case PointsActionPostCreated:
		return "post_created"
	default:
		return "unknown"
	}
}

func RewardPoints(userId uint64, points int64, action PointsAction) {
	_ = userPoints.Increment(userId, points)

	users.IncrementPrestige(points, userId)

	pointsRecordEntity := pointsRecord.Entity{
		UserId:       userId,
		Action:       action.Code(),
		PointsChange: points,
		CreatedAt:    time.Now(),
	}
	pointsRecord.Save(&pointsRecordEntity)

}

func InitUserPoints(userId uint64, points int64) {
	userPoint := userPoints.Get(userId)
	if userPoint.UserId > 0 {
		return
	}
	userPoint.UserId = userId
	userPoint.CurrentPoints += points
	userPoints.Create(&userPoint)

	pointsRecordEntity := pointsRecord.Entity{
		UserId:       userId,
		Action:       PointsActionInit.Code(),
		PointsChange: points,
		CreatedAt:    time.Now(),
	}
	pointsRecord.Save(&pointsRecordEntity)
}
