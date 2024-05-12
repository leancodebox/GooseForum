package optlogger

import (
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/spf13/cast"
	"time"
)

func UserOpt(userId uint64, optType int, targetType int, targetId any, msg string) {
	entity := optRecord.Entity{OptUserId: userId, OptType: optType, TargetType: targetType, TargetId: cast.ToString(targetId), OptInfo: msg, CreatedAt: time.Now()}
	optRecord.Create(&entity)
}
