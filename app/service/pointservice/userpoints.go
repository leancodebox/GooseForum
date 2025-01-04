package pointservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"time"
)

type RewardPointsType int

var (
	RewardPointsInit           RewardPointsType = 0
	RewardPoints4WriteArticles RewardPointsType = 1
	RewardPoints4Reply         RewardPointsType = 2
)

func (r RewardPointsType) String() string {
	switch r {
	case 0:
		return "初始化"
	case 1:
		return ""
	case 2:
		return ""
	}
	return ""
}

func RewardPoints(userId uint64, points int64, reason RewardPointsType) {
	userPoint := userPoints.Get(userId)
	userPoint.CurrentPoints += points
	userPoints.Save(&userPoint)

	users.IncrementPrestige(points, userId)

	pointsRecordEntity := pointsRecord.Entity{UserId: userId, ChangeReason: reason.String(), CreatedAt: time.Now()}
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

	pointsRecordEntity := pointsRecord.Entity{UserId: userId, ChangeReason: RewardPointsInit.String(), CreatedAt: time.Now()}
	pointsRecord.Save(&pointsRecordEntity)
}
