package jwtopt

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"github.com/spf13/cast"
	"strings"
	"sync"
	"time"
)

var (
	once       sync.Once
	std        *JWT
	signingKey = preferences.Get("jwtopt.signingKey", "mq+ZeGafL+b1xdC0u9vSVg==")
	validTime  = time.Duration(preferences.GetInt64("jwtopt.validTime", 86400*7)) * time.Second
)

func Std() *JWT {
	once.Do(func() {
		std = NewJWT([]byte(signingKey))
	})
	return std
}

func CreateNewToken(userId uint64, expireTime time.Duration) (string, error) {
	cc := CustomClaims{
		UserId:           userId,
		RegisteredClaims: GetBaseRegisteredClaims(expireTime),
	}
	return Std().CreateToken(cc)
}

// VerifyTokenWithFresh 验证token 并刷新， 如果token还有1天就过期则生成新的token，否则还是用原来的
func VerifyTokenWithFresh(tokenStr string) (userId uint64, newToken string, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, "", err
	}
	eTime, err := claims.GetExpirationTime()
	if err == nil && time.Now().Add(time.Second*86400*1).After(eTime.Time) {
		claims.RegisteredClaims = GetBaseRegisteredClaims(validTime)
		tokenStr, err = Std().CreateToken(*claims)
	}
	return claims.UserId, tokenStr, err
}

func VerifyToken(tokenStr string) (userId uint64, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.UserId, err
}

func GetGinAccessToken(c *gin.Context) string {
	var token string
	token = c.GetHeader("Authorization")
	token = strings.ReplaceAll(token, "Bearer ", "")
	if token == "" {
		token, _ = c.Cookie("access_token")
	}
	return token
}

func TokenSetting(c *gin.Context, newToken string) {
	c.Header("New-Token", newToken)
	c.SetCookie(
		"access_token",
		newToken,
		cast.ToInt(validTime/time.Second), // 24小时
		"/",
		"",    // 域名，为空表示当前域名
		false, // 仅HTTPS
		true,  // HttpOnly
	)
}

func TokenClean(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1, // 过期
		"/",
		"",    // 域名，为空表示当前域名
		false, // 仅HTTPS
		true,  // HttpOnly
	)
}
