package jwtopt

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/spf13/cast"
)

var (
	once       sync.Once
	std        *JWT
	signingKey = preferences.Get("app.signingKey", "mq+ZeGafL+b1xdC0u9vSVg==")
	validTime  = time.Duration(preferences.GetInt64("jwtopt.validTime", 86400*7)) * time.Second
)

// Std returns the process-wide JWT helper.
func Std() *JWT {
	once.Do(func() {
		std = NewJWT([]byte(signingKey))
	})
	return std
}

// CreateNewTokenDefault creates an access token with the configured lifetime.
func CreateNewTokenDefault(userId uint64) (string, error) {
	return CreateNewToken(userId, validTime)
}

// CreateNewToken creates an access token with expireTime.
func CreateNewToken(userId uint64, expireTime time.Duration) (string, error) {
	cc := CustomClaims{
		UserId:           userId,
		RegisteredClaims: GetBaseRegisteredClaims(expireTime),
	}
	return Std().CreateToken(cc)
}

// VerifyTokenWithFresh verifies tokenStr and refreshes it when it is close to expiry.
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

// VerifyToken verifies tokenStr and returns the user ID.
func VerifyToken(tokenStr string) (userId uint64, err error) {
	claims, err := Std().ParseToken(tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.UserId, err
}

// GetGinAccessToken returns the bearer token or access_token cookie from c.
func GetGinAccessToken(c *gin.Context) string {
	var token string
	token = c.GetHeader("Authorization")
	token = strings.ReplaceAll(token, "Bearer ", "")
	if token == "" {
		token, _ = c.Cookie("access_token")
	}
	return token
}

// TokenSetting writes the refreshed token to headers and cookies.
func TokenSetting(c *gin.Context, newToken string) {
	c.Header("New-Token", newToken)
	c.SetCookie(
		"access_token",
		newToken,
		cast.ToInt(validTime/time.Second),
		"/",
		"",
		false,
		true,
	)
}

// TokenClean expires the access_token cookie.
func TokenClean(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
}
