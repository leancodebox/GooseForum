package userservice

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

func CreateUser(username, password, email string, needValid bool) (*users.EntityComplete, error) {
	userEntity := users.MakeUser(username, password, email)
	userEntity.Nickname = GenerateGooseNickname()
	if !needValid {
		userEntity.IsActivated = users.ActivationSuccess
	}
	// 默认状态为正常
	userEntity.IsFrozen = users.StatusNormal
	// 初始化用户积分
	pointservice.InitUserPoints(userEntity.Id, 100)
	err := users.Create(userEntity)
	if err != nil {
		return nil, err
	}
	userSt := userStatistics.Entity{UserId: userEntity.Id}
	// 初始化统计表
	userStatistics.SaveOrCreateById(&userSt)
	if userEntity.Id == 1 {
		// For the first user registered, elevate it to admin group.
		FirstUserInit(userEntity)
	}
	return userEntity, nil
}

// GenerateGooseNickname 生成鹅相关昵称（优化长度版）
func GenerateGooseNickname() string {
	prefixes := []string{
		"鹅", "大白鹅", "灰鹅", "小鹅", "鹅宝",
		"Goose", "Gander", "Gosling", "Honker",
	}
	prefix := prefixes[rand.Intn(len(prefixes))]

	// 使用毫秒级时间戳+随机数确保唯一性
	now := time.Now()
	timestamp := now.UnixMilli()  // 毫秒级时间戳
	randomPart := rand.Intn(1296) // 0-1295 (36^2范围)

	// 转换为36进制（0-9a-z）
	timestamp36 := strconv.FormatInt(timestamp, 36)
	random36 := strconv.FormatInt(int64(randomPart), 36)

	// 确保随机数部分固定2位（补零）
	if len(random36) < 2 {
		random36 = "0" + random36
	}

	// 组合成更紧凑的字符串
	return fmt.Sprintf("%s%s%s", prefix, timestamp36, random36)
}
