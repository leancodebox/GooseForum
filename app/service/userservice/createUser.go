package userservice

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

func CreateUser(username, password, email string, needValid bool, locale ...string) (*users.EntityComplete, error) {
	userEntity := users.MakeUser(username, password, email)
	userEntity.Locale = normalizeUserLocale(locale...)
	userEntity.Nickname = GenerateGooseNickname()
	if !needValid {
		userEntity.IsActivated = users.ActivationSuccess
	}
	userEntity.IsFrozen = users.StatusNormal
	pointservice.InitUserPoints(userEntity.Id, 100)
	err := users.Create(userEntity)
	if err != nil {
		return nil, err
	}
	userSt := userStatistics.Entity{UserId: userEntity.Id}
	userStatistics.SaveOrCreateById(&userSt)
	if userEntity.Id == 1 {
		FirstUserInit(userEntity)
	}
	return userEntity, nil
}

func normalizeUserLocale(values ...string) string {
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return ""
	}
	return i18n.Normalize(values[0])
}

// GenerateGooseNickname creates a compact random default nickname.
func GenerateGooseNickname() string {
	prefixes := []string{
		"鹅", "大白鹅", "灰鹅", "小鹅", "鹅宝",
		"Goose", "Gander", "Gosling", "Honker",
	}
	prefix := prefixes[rand.Intn(len(prefixes))]

	now := time.Now()
	timestamp := now.UnixMilli()
	randomPart := rand.Intn(1296)

	timestamp36 := strconv.FormatInt(timestamp, 36)
	random36 := strconv.FormatInt(int64(randomPart), 36)

	if len(random36) < 2 {
		random36 = "0" + random36
	}

	return fmt.Sprintf("%s%s%s", prefix, timestamp36, random36)
}
