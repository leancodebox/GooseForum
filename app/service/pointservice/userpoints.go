package points

import (
	"github.com/leancodebox/GooseForum/app/models/forum/pointsRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/userPoints"
	"time"
)

func RewardPoints(userId uint64, points int64, reason string) {
	userPoint := userPoints.Get(userId)
	userPoint.CurrentPoints += points
	userPoints.Save(&userPoint)

	pointsRecordEntity := pointsRecord.Entity{UserId: userId, ChangeReason: reason, CreatedAt: time.Now()}
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

	pointsRecordEntity := pointsRecord.Entity{UserId: userId, ChangeReason: "用户创建奖励", CreatedAt: time.Now()}
	pointsRecord.Save(&pointsRecordEntity)
}
