package component

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
	"regexp"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{6,32}$`)
)

func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func LoginUserId(c *gin.Context) uint64 {
	return c.GetUint64("userId")
}

func GetLoginUser(c *gin.Context) *vo.UserInfoShow {
	userId := LoginUserId(c)
	return GetUserShowByUserId(userId)
}

func GetUserShowByUserId(userId uint64) *vo.UserInfoShow {
	if userId == 0 {
		return &vo.UserInfoShow{}
	}
	return hotdataserve.GetOrLoad(fmt.Sprintf("user:%v", userId), func() (*vo.UserInfoShow, error) {
		user, _ := users.Get(userId)
		if user.Id == 0 {
			return &vo.UserInfoShow{}, errors.New("no found")
		}
		return transform.User2userShow(user), nil
	})
}

func SendAEmail4User(userEntity *users.EntityComplete) error {
	token, err := tokenservice.GenerateActivationTokenByUser(*userEntity)
	if err != nil {
		return err
	}

	// 将邮件任务加入队列
	err = mailservice.AddToQueue(mailservice.EmailTask{
		To:       userEntity.Email,
		Username: userEntity.Username,
		Token:    token,
		Type:     "activation",
	})
	if err != nil {
		return nil
	}
	return nil
}
