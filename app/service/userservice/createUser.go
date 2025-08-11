package userservice

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

func CreateUser(username, password, email string, needValid bool) (*users.EntityComplete, error) {
	userEntity := users.MakeUser(username, password, email)
	userEntity.Nickname = GenerateGooseNickname()
	if !needValid {
		userEntity.Validate = 1
	}
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

// GenerateGooseNickname 新增生成鹅相关昵称的函数
func GenerateGooseNickname() string {
	prefixes := []string{
		"鹅", "大白鹅", "灰鹅", "小鹅", "鹅宝",
		"Goose", "Gander", "Gosling", "Honker",
	}
	prefix := prefixes[rand.Intn(len(prefixes))]
	// 使用纳秒级时间戳+随机数确保唯一性
	now := time.Now()
	timestamp := now.UnixNano()
	randomPart := rand.Intn(1000)
	// 组合成16进制字符串
	return fmt.Sprintf("%s%x%03d", prefix, timestamp, randomPart)
}
